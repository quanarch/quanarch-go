// Package host is the quan's own program. The quanar's quan-build links it with a quan's
// package: this is the main the quan does not have. It plugs shop-qua in, tells the quanar what
// the quan declared and runs the quan's handlers, middlewares and subscriptions.
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

	shopqua "shop-qua"
)

// message is one line between the quanar and the quan.
type message struct {
	Type         string               `json:"type"`
	ID           int64                `json:"id,omitempty"`
	Quas         []string             `json:"quas,omitempty"`
	Api          int                  `json:"api,omitempty"`
	Req          json.RawMessage      `json:"req,omitempty"`
	Res          *shopqua.ApiResponse `json:"res,omitempty"`
	Subscription int                  `json:"subscription,omitempty"`
	Payload      json.RawMessage      `json:"payload,omitempty"`
	Meta         *shopqua.EventMeta   `json:"meta,omitempty"`
	Capability   string               `json:"capability,omitempty"`
	Args         []any                `json:"args,omitempty"`
	Value        json.RawMessage      `json:"value,omitempty"`
	Error        string               `json:"error,omitempty"`
	Manifest     any                  `json:"manifest,omitempty"`
	Level        string               `json:"level,omitempty"`
	Message      string               `json:"message,omitempty"`
	Data         any                  `json:"data,omitempty"`
}

var (
	out         *os.File
	writing     sync.Mutex
	callTimeout = 5 * time.Second
	manifest    shopqua.QuanManifest
)

func send(m message) {
	line, err := json.Marshal(m)
	if err != nil {
		// What the quan answered cannot travel; tell the quanar so nothing hangs.
		line, _ = json.Marshal(message{Type: m.Type, ID: m.ID, Error: "cannot send " + m.Type + ": " + err.Error()})
	}
	writing.Lock()
	defer writing.Unlock()
	out.Write(append(line, '\n'))
}

// Run is the quan's program: m is the quan's Manifest.
func Run(m shopqua.QuanManifest) {
	flag.DurationVar(&callTimeout, "call-timeout", callTimeout, "how long the quan waits for the quanar")
	flag.Parse()
	manifest = m

	// The protocol keeps the real stdout; whatever the quan prints goes to stderr.
	out = os.Stdout
	os.Stdout = os.Stderr

	in := bufio.NewReaderSize(os.Stdin, 1<<20)
	for {
		line, err := in.ReadBytes('\n')
		if err != nil {
			return // the quanar is gone
		}
		var msg message
		if json.Unmarshal(line, &msg) != nil {
			continue
		}
		// Nothing here waits on the quan: its calls are answered on this same stream.
		switch msg.Type {
		case "load":
			load(msg.Quas)
		case "start":
			go start()
		case "request":
			go handle(msg)
		case "event":
			go deliver(msg)
		case "result":
			settle(msg)
		case "unmount":
			if manifest.OnUnmount != nil {
				_ = manifest.OnUnmount(context.Background())
			}
			return
		}
	}
}

// apiMeta is an api as the quanar keeps it: everything but the code.
type apiMeta struct {
	Endpoint         string                   `json:"endpoint"`
	Method           string                   `json:"method"`
	Middlewares      []string                 `json:"middlewares"`
	Permissions      []string                 `json:"permissions"`
	Validation       *shopqua.ValidationRules `json:"validation"`
	Description      string                   `json:"description"`
	LocalMiddlewares int                      `json:"local_middlewares"`
}

func load(quas []string) {
	defer func() {
		if r := recover(); r != nil {
			send(message{Type: "load-failed", Error: fmt.Sprint(r)})
		}
	}()
	for _, qua := range quas {
		if qua == "shop-qua" {
			if err := shopqua.Init(core{}); err != nil {
				send(message{Type: "load-failed", Error: err.Error()})
				return
			}
		}
	}
	apis := []apiMeta{}
	for _, a := range manifest.Apis {
		middlewares := []string{}
		for _, m := range a.QuanarMiddlewares {
			middlewares = append(middlewares, string(m))
		}
		permissions := a.Permissions
		if permissions == nil {
			permissions = []string{}
		}
		apis = append(apis, apiMeta{a.Endpoint, string(a.Method), middlewares, permissions, a.Validation, a.Description, len(a.Middlewares)})
	}
	events := []string{}
	for _, e := range manifest.Events {
		events = append(events, e.Event)
	}
	models, permissions := manifest.Models, manifest.Permissions
	if models == nil {
		models = []shopqua.ModelDefinition{}
	}
	if permissions == nil {
		permissions = []shopqua.PermissionDefinition{}
	}
	send(message{Type: "loaded", Manifest: map[string]any{
		"name": manifest.Name, "version": manifest.Version, "models": models,
		"permissions": permissions, "apis": apis, "events": events,
	}})
}

func start() {
	if manifest.OnMount != nil {
		if err := manifest.OnMount(context.Background()); err != nil {
			send(message{Type: "started", Error: err.Error()})
			return
		}
	}
	send(message{Type: "started"})
}

func handle(msg message) {
	res, err := func() (res shopqua.ApiResponse, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()
		var req shopqua.ApiRequest
		if err := json.Unmarshal(msg.Req, &req); err != nil {
			return res, err
		}
		req.Locals = map[string]any{}
		api := manifest.Apis[msg.Api]
		ctx := context.Background()
		// The quan's own middlewares, then its handler. A middleware may answer without calling next.
		var run func(i int) (shopqua.ApiResponse, error)
		run = func(i int) (shopqua.ApiResponse, error) {
			if i < len(api.Middlewares) {
				return api.Middlewares[i](ctx, &req, func() (shopqua.ApiResponse, error) { return run(i + 1) })
			}
			return api.Handler(ctx, &req)
		}
		return run(0)
	}()
	if err != nil {
		send(message{Type: "response", ID: msg.ID, Error: err.Error()})
		return
	}
	send(message{Type: "response", ID: msg.ID, Res: &res})
}

func deliver(msg message) {
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()
		return manifest.Events[msg.Subscription].Handler(context.Background(), msg.Payload, *msg.Meta)
	}()
	if err != nil {
		send(message{Type: "delivered", ID: msg.ID, Error: err.Error()})
		return
	}
	send(message{Type: "delivered", ID: msg.ID})
}
