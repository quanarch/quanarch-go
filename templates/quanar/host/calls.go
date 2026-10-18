package host

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	pendingMu sync.Mutex
	nextID    int64
	pending   = map[int64]chan message{}
)

// call asks the quanar for a capability and waits for its answer, at most callTimeout.
func call(ctx context.Context, capability string, out any, args ...any) error {
	pendingMu.Lock()
	nextID++
	id := nextID
	answer := make(chan message, 1)
	pending[id] = answer
	pendingMu.Unlock()
	defer func() {
		pendingMu.Lock()
		delete(pending, id)
		pendingMu.Unlock()
	}()

	send(message{Type: "call", ID: id, Capability: capability, Args: args})
	select {
	case m := <-answer:
		if m.Error != "" {
			return errors.New(m.Error)
		}
		if out == nil || len(m.Value) == 0 {
			return nil
		}
		return json.Unmarshal(m.Value, out)
	case <-time.After(callTimeout):
		return fmt.Errorf("the quanar did not answer %s in %s", capability, callTimeout)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func settle(m message) {
	pendingMu.Lock()
	answer, ok := pending[m.ID]
	pendingMu.Unlock()
	if ok {
		answer <- m
	}
}
