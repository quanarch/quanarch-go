// Package events carries events between quans. An emitter does not know who listens; a listener
// does not know who emits. Delivery is at most once: a quan that is not mounted misses it.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"shop-quanar/internal/wire"
)

// Deliver hands an event to the subscription at index in a quan.
type Deliver func(ctx context.Context, subscription int, payload json.RawMessage, meta wire.Meta) error

type listener struct {
	events  []string
	deliver Deliver
}

// Manager carries events.
type Manager struct {
	mu        sync.RWMutex
	listeners map[string]listener
}

// New makes an event manager.
func New() *Manager { return &Manager{listeners: map[string]listener{}} }

// Subscribe keeps what a quan listens to.
func (e *Manager) Subscribe(owner string, events []string, deliver Deliver) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners[owner] = listener{events, deliver}
}

// Forget drops a quan's subscriptions.
func (e *Manager) Forget(owner string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.listeners, owner)
}

// Publish carries an event to every listener and returns how many it reached. Its name starts
// with the emitting quan's name.
func (e *Manager) Publish(emittedBy, event string, payload json.RawMessage) (int, error) {
	if !strings.HasPrefix(event, emittedBy+".") {
		return 0, fmt.Errorf("%s may not emit %s: events start with %s.", emittedBy, event, emittedBy)
	}
	meta := wire.Meta{EmittedBy: emittedBy, At: time.Now().UTC().Format("2006-01-02T15:04:05.000Z")}
	reached := 0
	e.mu.RLock()
	for owner, l := range e.listeners {
		for subscription, name := range l.events {
			if name != event {
				continue
			}
			reached++
			go func(owner string, deliver Deliver, subscription int) {
				if err := deliver(context.Background(), subscription, payload, meta); err != nil {
					slog.Error(owner+" failed on "+event, "error", err, "scope", "events")
				}
			}(owner, l.deliver, subscription)
		}
	}
	e.mu.RUnlock()
	suffix := "s"
	if reached == 1 {
		suffix = ""
	}
	slog.Info(fmt.Sprintf("%s from %s → %d listener%s", event, emittedBy, reached, suffix), "scope", "events")
	return reached, nil
}

// Describe lists what each quan listens to.
func (e *Manager) Describe() map[string][]string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := map[string][]string{}
	for owner, l := range e.listeners {
		out[owner] = l.events
	}
	return out
}
