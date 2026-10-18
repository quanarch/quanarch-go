// Package process runs one quan in its own process. The quanar starts it through the gate, with
// an empty environment, and talks to it only through JSON lines.
package process

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"shop-quanar/gate"
	"shop-quanar/internal/apierr"
	"shop-quanar/internal/config"
	"shop-quanar/internal/quacore"
	"shop-quanar/internal/wire"
)

type message struct {
	Type         string            `json:"type"`
	ID           int64             `json:"id,omitempty"`
	Quas         []string          `json:"quas,omitempty"`
	Api          int               `json:"api,omitempty"`
	Req          *wire.Request     `json:"req,omitempty"`
	Res          *wire.Response    `json:"res,omitempty"`
	Subscription int               `json:"subscription,omitempty"`
	Payload      json.RawMessage   `json:"payload,omitempty"`
	Meta         *wire.Meta        `json:"meta,omitempty"`
	Capability   string            `json:"capability,omitempty"`
	Args         []json.RawMessage `json:"args,omitempty"`
	Value        any               `json:"value,omitempty"`
	Error        string            `json:"error,omitempty"`
	Manifest     *wire.Manifest    `json:"manifest,omitempty"`
	Level        string            `json:"level,omitempty"`
	Message      string            `json:"message,omitempty"`
	Data         json.RawMessage   `json:"data,omitempty"`
}

// Process is one running quan.
type Process struct {
	source   config.QuanSource
	settings config.Settings
	log      *slog.Logger
	caps     map[string]quacore.Capability
	dir      string
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	writing  sync.Mutex
	mu       sync.Mutex
	nextID   int64
	waiting  map[int64]chan message
	once     map[string]chan message
	exited   chan struct{}
	// OnExit runs once the process is gone.
	OnExit func()
}

// Start writes the quan's program to a private folder, runs it through the gate and loads it.
func Start(ctx context.Context, source config.QuanSource, settings config.Settings, core quacore.Core, program []byte) (*Process, wire.Manifest, error) {
	var manifest wire.Manifest
	dir, err := os.MkdirTemp("", "quan-"+source.Name+"-")
	if err != nil {
		return nil, manifest, err
	}
	path := filepath.Join(dir, source.Name)
	if err := os.WriteFile(path, program, 0o500); err != nil {
		os.RemoveAll(dir)
		return nil, manifest, err
	}
	self, err := os.Executable()
	if err != nil {
		os.RemoveAll(dir)
		return nil, manifest, err
	}
	p := &Process{
		source: source, settings: settings, dir: dir,
		log:     slog.With("quan", source.Name),
		caps:    core.Plugged(source.Name, source.Quas),
		waiting: map[int64]chan message{},
		once:    map[string]chan message{"loaded": make(chan message, 1), "started": make(chan message, 1)},
		exited:  make(chan struct{}),
	}
	var name string
	var args []string
	if gate.Walled {
		name, args = self, []string{gate.Command, "--memory-mb", fmt.Sprint(settings.QuanMemoryMB), "--call-timeout", settings.CallTimeout.String(), path}
	} else {
		name, args = path, []string{"--call-timeout", settings.CallTimeout.String()}
	}
	stdout, err := p.launch(name, args)
	if err != nil {
		os.RemoveAll(dir)
		return nil, manifest, err
	}
	go p.read(stdout)

	m, err := p.answerOnce(ctx, "loaded", message{Type: "load", Quas: source.Quas})
	if err != nil {
		return nil, manifest, err
	}
	// The program runs now; its file is not needed. Where a running program's file cannot be
	// removed, it goes when the process ends.
	_ = os.RemoveAll(dir)
	return p, *m.Manifest, nil
}

// Begin lets the quan run its OnMount, once its models and apis are in place.
func (p *Process) Begin(ctx context.Context) error {
	_, err := p.answerOnce(ctx, "started", message{Type: "start"})
	return err
}

func (p *Process) answerOnce(ctx context.Context, kind string, m message) (message, error) {
	p.send(m)
	select {
	case answer := <-p.once[kind]:
		if answer.Error != "" {
			p.Kill()
			return answer, errors.New(answer.Error)
		}
		return answer, nil
	case <-p.exited:
		return message{}, fmt.Errorf("%s exited before %s", p.source.Name, kind)
	case <-time.After(p.settings.LoadTimeout):
		p.Kill()
		return message{}, fmt.Errorf("%s did not answer %s in %s", p.source.Name, m.Type, p.settings.LoadTimeout)
	case <-ctx.Done():
		p.Kill()
		return message{}, ctx.Err()
	}
}

// launch starts the quan's program the best way the system allows.
func (p *Process) launch(name string, args []string) (io.Reader, error) {
	var err error
	for i, attr := range sysProcAttrs() {
		if i > 0 {
			fellBack(err)
		}
		p.cmd = exec.Command(name, args...)
		p.cmd.Env = []string{} // no secrets, no addresses, nothing of the quanar's
		p.cmd.Stderr = os.Stderr
		p.cmd.SysProcAttr = attr
		if p.stdin, err = p.cmd.StdinPipe(); err != nil {
			return nil, err
		}
		stdout, perr := p.cmd.StdoutPipe()
		if perr != nil {
			return nil, perr
		}
		if err = p.cmd.Start(); err == nil {
			return stdout, nil
		}
	}
	return nil, err
}

// PID is the quan's process id.
func (p *Process) PID() int { return p.cmd.Process.Pid }

func (p *Process) ask(ctx context.Context, m message) (message, error) {
	p.mu.Lock()
	p.nextID++
	id := p.nextID
	answer := make(chan message, 1)
	p.waiting[id] = answer
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		delete(p.waiting, id)
		p.mu.Unlock()
	}()
	m.ID = id
	p.send(m)
	select {
	case a := <-answer:
		return a, nil
	case <-p.exited:
		return message{}, apierr.New(http.StatusServiceUnavailable, apierr.QuanUnavailable, p.source.Name+" stopped")
	case <-time.After(p.settings.RequestTimeout):
		return message{}, context.DeadlineExceeded
	case <-ctx.Done():
		return message{}, ctx.Err()
	}
}

// Request hands a request to the api at index in the quan.
func (p *Process) Request(ctx context.Context, index int, req wire.Request) (wire.Response, error) {
	a, err := p.ask(ctx, message{Type: "request", Api: index, Req: &req})
	if errors.Is(err, context.DeadlineExceeded) {
		// A quan stuck in a loop is shut down; its process goes, the quanar stays.
		p.log.Error("no answer, stopping it", "after", p.settings.RequestTimeout)
		p.Kill()
		return wire.Response{}, apierr.New(http.StatusGatewayTimeout, apierr.QuanTimeout, p.source.Name+" did not answer")
	}
	if err != nil {
		return wire.Response{}, err
	}
	if a.Error != "" || a.Res == nil {
		p.log.Error(req.Method+" "+req.Path+" failed", "error", a.Error)
		return wire.Response{}, apierr.New(http.StatusInternalServerError, apierr.QuanFailed, p.source.Name+" failed")
	}
	return *a.Res, nil
}

// Deliver hands an event to the subscription at index in the quan.
func (p *Process) Deliver(ctx context.Context, subscription int, payload json.RawMessage, meta wire.Meta) error {
	a, err := p.ask(ctx, message{Type: "event", Subscription: subscription, Payload: payload, Meta: &meta})
	if err != nil {
		return err
	}
	if a.Error != "" {
		return errors.New(a.Error)
	}
	return nil
}

// Stop asks the quan to clean up, then makes sure its process is gone.
func (p *Process) Stop() {
	p.send(message{Type: "unmount"})
	select {
	case <-p.exited:
	case <-time.After(2 * time.Second):
		p.Kill()
		<-p.exited
	}
}

// Kill ends the quan's process at once.
func (p *Process) Kill() { _ = p.cmd.Process.Kill() }

func (p *Process) send(m message) {
	line, err := json.Marshal(m)
	if err != nil {
		p.log.Error("cannot send to the quan", "type", m.Type, "error", err)
		return
	}
	p.writing.Lock()
	defer p.writing.Unlock()
	_, _ = p.stdin.Write(append(line, '\n'))
}

func (p *Process) read(stdout io.Reader) {
	lines := bufio.NewReaderSize(stdout, 1<<20)
	for {
		line, err := lines.ReadBytes('\n')
		if err != nil {
			break
		}
		var m message
		if json.Unmarshal(line, &m) != nil {
			continue
		}
		switch m.Type {
		case "loaded", "started":
			p.once[m.Type] <- m
		case "load-failed":
			p.once["loaded"] <- message{Type: "loaded", Error: m.Error}
		case "response", "delivered":
			p.mu.Lock()
			answer := p.waiting[m.ID]
			p.mu.Unlock()
			if answer != nil {
				answer <- m
			}
		case "call":
			go p.answer(m)
		case "log":
			level := slog.LevelInfo
			switch m.Level {
			case "warn":
				level = slog.LevelWarn
			case "error":
				level = slog.LevelError
			}
			if len(m.Data) > 0 && string(m.Data) != "null" {
				p.log.Log(context.Background(), level, m.Message, "data", json.RawMessage(m.Data))
			} else {
				p.log.Log(context.Background(), level, m.Message)
			}
		}
	}
	_ = p.cmd.Wait()
	os.RemoveAll(p.dir)
	close(p.exited)
	if p.OnExit != nil {
		p.OnExit()
	}
}

func (p *Process) answer(m message) {
	capability, ok := p.caps[m.Capability]
	if !ok {
		p.send(message{Type: "result", ID: m.ID, Error: m.Capability + " is not plugged into " + p.source.Name})
		return
	}
	value, err := capability(context.Background(), m.Args)
	if err != nil {
		p.send(message{Type: "result", ID: m.ID, Error: err.Error()})
		return
	}
	p.send(message{Type: "result", ID: m.ID, Value: value})
}
