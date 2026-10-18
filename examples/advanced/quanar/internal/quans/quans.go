// Package quans follows each quan's event stream: load on open, reload on change, unmount on close.
package quans

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"shop-quanar/internal/config"
	"shop-quanar/internal/eventstream"
	"shop-quanar/internal/loader"
	"shop-quanar/internal/registry"
)

// Status is a quan as /quans shows it.
type Status struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Quas    []string `json:"quas"`
	State   string   `json:"state"`
	Detail  string   `json:"detail,omitempty"`
	PID     int      `json:"pid,omitempty"`
}

// Service follows the quans of a quanar.
type Service struct {
	settings config.Settings
	sources  []config.QuanSource
	registry *registry.Registry
	mu       sync.Mutex
	states   map[string][2]string
	locks    map[string]*sync.Mutex
}

// New makes a service for the quans the settings list.
func New(settings config.Settings, r *registry.Registry) *Service {
	s := &Service{settings: settings, sources: settings.QuanSources(), registry: r, states: map[string][2]string{}, locks: map[string]*sync.Mutex{}}
	for _, source := range s.sources {
		s.states[source.Name] = [2]string{"pending", ""}
		s.locks[source.Name] = &sync.Mutex{}
	}
	return s
}

func (s *Service) setState(name, state, detail string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.states[name][0] != state && state != "mounted" {
		slog.Warn(name+": "+state, "detail", detail, "scope", "quans")
	}
	s.states[name] = [2]string{state, detail}
}

func (s *Service) sync(ctx context.Context, source config.QuanSource) {
	lock := s.locks[source.Name]
	lock.Lock()
	defer lock.Unlock()
	fetchCtx, cancel := context.WithTimeout(ctx, s.settings.FetchTimeout)
	program, err := loader.Fetch(fetchCtx, source)
	cancel()
	var hash string
	if err == nil {
		hash, err = loader.Verify(source, program)
	}
	if err != nil {
		state := "unreachable"
		if errors.Is(err, loader.ErrHashMismatch) {
			state = "hash-mismatch"
		}
		s.setState(source.Name, state, err.Error())
		s.registry.Unmount(source.Name, state)
		return
	}
	if s.registry.Hash(source.Name) == hash {
		return
	}
	if err := s.registry.Mount(ctx, source, program, hash, func() { s.setState(source.Name, "crashed", "process exited") }); err != nil {
		s.setState(source.Name, "failed", err.Error())
		s.registry.Unmount(source.Name, "failed")
		return
	}
	s.setState(source.Name, "mounted", "")
}

func (s *Service) watch(ctx context.Context, source config.QuanSource) {
	if source.Events == "" {
		s.sync(ctx, source)
		return
	}
	for ctx.Err() == nil {
		eventstream.Follow(ctx, source.Events,
			func() { go s.sync(ctx, source) },
			func(name string) {
				if name == "change" {
					go s.sync(ctx, source)
				}
			})
		s.setState(source.Name, "unreachable", source.Events+" closed")
		s.registry.Unmount(source.Name, "its event stream closed")
		select {
		case <-ctx.Done():
		case <-time.After(s.settings.Reconnect):
		}
	}
}

// Start follows every quan until ctx ends.
func (s *Service) Start(ctx context.Context) {
	if len(s.sources) == 0 {
		slog.Warn("no quans listed in "+s.settings.QuansFile, "scope", "quans")
	}
	for _, source := range s.sources {
		go s.watch(ctx, source)
	}
}

// Stop unmounts every quan.
func (s *Service) Stop() {
	for _, name := range s.registry.Names() {
		s.registry.Unmount(name, "quanar stopping")
	}
}

// Statuses lists every quan's state.
func (s *Service) Statuses() []Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Status{}
	for _, source := range s.sources {
		version := s.registry.Version(source.Name)
		if version == "" {
			version = source.Version
		}
		out = append(out, Status{source.Name, version, source.Quas, s.states[source.Name][0], s.states[source.Name][1], s.registry.PID(source.Name)})
	}
	return out
}
