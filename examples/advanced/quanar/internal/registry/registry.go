// Package registry turns what a quan declared into a running quan, through the managers: its
// models become tables, its permissions become known, its apis open, its subscriptions start
// receiving events, and only then does it start.
package registry

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"shop-quanar/internal/apis"
	"shop-quanar/internal/config"
	"shop-quanar/internal/events"
	"shop-quanar/internal/permissions"
	"shop-quanar/internal/process"
	"shop-quanar/internal/quacore"
	"shop-quanar/internal/schema"
)

type mounted struct {
	hash    string
	version string
	process *process.Process
}

// Registry keeps the mounted quans.
type Registry struct {
	settings    config.Settings
	core        quacore.Core
	schema      *schema.Manager
	permissions *permissions.Manager
	apis        *apis.Manager
	events      *events.Manager
	mu          sync.Mutex
	mounted     map[string]mounted
}

// New makes a registry over the managers.
func New(settings config.Settings, core quacore.Core, s *schema.Manager, p *permissions.Manager, a *apis.Manager, e *events.Manager) *Registry {
	return &Registry{settings: settings, core: core, schema: s, permissions: p, apis: a, events: e, mounted: map[string]mounted{}}
}

// Mount starts a quan's program and makes what it declared real. onCrash runs if it dies later.
func (r *Registry) Mount(ctx context.Context, source config.QuanSource, program []byte, hash string, onCrash func()) error {
	p, manifest, err := process.Start(ctx, source, r.settings, r.core, program)
	if err != nil {
		return err
	}
	fail := func(err error) error {
		p.Kill()
		r.mu.Lock()
		current, ok := r.mounted[source.Name]
		r.mu.Unlock()
		if ok && current.process == p {
			r.Release(source.Name)
		}
		return err
	}
	if manifest.Name != source.Name {
		return fail(fmt.Errorf("the manifest says %q, the list says %q", manifest.Name, source.Name))
	}

	r.Unmount(source.Name, manifest.Version+" arrived")

	if err := r.schema.Sync(source.Name, manifest.Models); err != nil {
		return fail(err)
	}
	if err := r.permissions.Register(source.Name, manifest.Permissions); err != nil {
		return fail(err)
	}
	if err := r.apis.Create(source.Name, manifest.Apis, p.Request); err != nil {
		return fail(err)
	}
	r.events.Subscribe(source.Name, manifest.Events, p.Deliver)
	r.mu.Lock()
	r.mounted[source.Name] = mounted{hash, manifest.Version, p}
	r.mu.Unlock()
	if err := p.Begin(ctx); err != nil {
		return fail(err)
	}
	p.OnExit = func() {
		r.mu.Lock()
		current, ok := r.mounted[source.Name]
		r.mu.Unlock()
		if ok && current.process == p {
			r.Release(source.Name)
			onCrash()
		}
	}
	slog.Info(fmt.Sprintf("%s@%s mounted in pid %d: %d apis, %d models, %d subscriptions",
		source.Name, manifest.Version, p.PID(), len(manifest.Apis), len(manifest.Models), len(manifest.Events)), "scope", "registry")
	return nil
}

// Unmount stops a quan and takes it out of every manager.
func (r *Registry) Unmount(name, reason string) {
	r.mu.Lock()
	quan, ok := r.mounted[name]
	r.mu.Unlock()
	if !ok {
		return
	}
	r.Release(name)
	quan.process.Stop()
	slog.Warn(name+" unmounted: "+reason, "scope", "registry")
}

// Release takes a quan out of every manager. Its tables and rows stay.
func (r *Registry) Release(name string) {
	r.apis.Remove(name)
	r.events.Forget(name)
	r.permissions.Remove(name)
	r.schema.Remove(name)
	r.mu.Lock()
	delete(r.mounted, name)
	r.mu.Unlock()
}

// Hash is the hash of a mounted quan's program.
func (r *Registry) Hash(name string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.mounted[name].hash
}

// Version is the version a mounted quan declared.
func (r *Registry) Version(name string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.mounted[name].version
}

// PID is a mounted quan's process id, or 0.
func (r *Registry) PID(name string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.mounted[name]; ok {
		return m.process.PID()
	}
	return 0
}

// Names lists the mounted quans.
func (r *Registry) Names() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	names := make([]string, 0, len(r.mounted))
	for n := range r.mounted {
		names = append(names, n)
	}
	return names
}
