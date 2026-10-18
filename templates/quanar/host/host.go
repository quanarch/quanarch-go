// Package host is the quan's own program. The quanar's quan-build links it with a quan's
// package: this is the main the quan does not have. It plugs the quas in, calls the quan's
// Mount and talks to the quanar in JSON lines over stdin and stdout.
package host

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

// message is one line between the quanar and the quan.
type message struct {
	Type       string          `json:"type"`
	ID         int64           `json:"id,omitempty"`
	Quas       []string        `json:"quas,omitempty"`
	Listener   int             `json:"listener,omitempty"`
	Req        json.RawMessage `json:"req,omitempty"`
	Res        any             `json:"res,omitempty"`
	Capability string          `json:"capability,omitempty"`
	Args       []any           `json:"args,omitempty"`
	Value      json.RawMessage `json:"value,omitempty"`
	Error      string          `json:"error,omitempty"`
	Listens    []listen        `json:"listens,omitempty"`
}

type listen struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

var (
	out         *os.File
	writing     sync.Mutex
	callTimeout = 5 * time.Second
)

func send(m message) {
	line, err := json.Marshal(m)
	if err != nil {
		// What the quan answered cannot travel; tell the quanar so the request does not hang.
		line, _ = json.Marshal(message{Type: m.Type, ID: m.ID, Error: "cannot send " + m.Type + ": " + err.Error()})
	}
	writing.Lock()
	defer writing.Unlock()
	out.Write(append(line, '\n'))
}

// Run is the quan's program: mount is the quan's Mount.
func Run(mount func()) {
	flag.DurationVar(&callTimeout, "call-timeout", callTimeout, "how long the quan waits for the quanar")
	flag.Parse()

	// The protocol keeps the real stdout; whatever the quan prints goes to stderr.
	out = os.Stdout
	os.Stdout = os.Stderr

	in := bufio.NewReaderSize(os.Stdin, 1<<20)
	for {
		line, err := in.ReadBytes('\n')
		if err != nil {
			return // the quanar is gone
		}
		var m message
		if json.Unmarshal(line, &m) != nil {
			continue
		}
		switch m.Type {
		case "load":
			load(m.Quas, mount)
		case "request":
			go handle(m)
		case "result":
			settle(m)
		case "unmount":
			return
		}
	}
}

func load(quas []string, mount func()) {
	defer func() {
		if r := recover(); r != nil {
			send(message{Type: "load-failed", Error: fmt.Sprint(r)})
		}
	}()
	for _, qua := range quas {
		if plug, ok := plugs[qua]; ok {
			if err := plug(); err != nil {
				send(message{Type: "load-failed", Error: err.Error()})
				return
			}
		}
	}
	mount()
	send(message{Type: "loaded", Listens: listens})
}

func handle(m message) {
	res, err := func() (res any, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()
		return handlers[m.Listener](context.Background(), m.Req)
	}()
	if err != nil {
		send(message{Type: "response", ID: m.ID, Error: err.Error()})
		return
	}
	send(message{Type: "response", ID: m.ID, Res: res})
}
