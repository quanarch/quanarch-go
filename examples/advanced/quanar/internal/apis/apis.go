// Package apis opens the apis quans declare. For each api, in one pass: the quanar's middlewares,
// then authentication and permissions, then validation, and only then the quan. Removing a quan
// closes its apis at once.
package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"shop-quanar/internal/apierr"
	"shop-quanar/internal/middleware"
	"shop-quanar/internal/wire"
)

// Invoke hands a request to the api at index in the quan.
type Invoke func(ctx context.Context, index int, req wire.Request) (wire.Response, error)

// Open is one api, open in the quanar.
type Open struct {
	Owner    string
	Index    int
	Meta     wire.Api
	segments []string
	steps    []middleware.Step
	invoke   Invoke
}

// Manager keeps every quan's open apis.
type Manager struct {
	mu        sync.RWMutex
	open      map[string][]*Open
	order     []string
	window    time.Duration
	rateLimit int
}

// New makes an api manager with the quanar's rate limit.
func New(window time.Duration, rateLimit int) *Manager {
	return &Manager{open: map[string][]*Open{}, window: window, rateLimit: rateLimit}
}

// Create opens a quan's apis; every endpoint starts with the quan's name.
func (m *Manager) Create(owner string, apis []wire.Api, invoke Invoke) error {
	var opened []*Open
	for index, api := range apis {
		if api.Endpoint != "/"+owner && !strings.HasPrefix(api.Endpoint, "/"+owner+"/") {
			return fmt.Errorf("%s may not listen to %s: endpoints start with /%s", owner, api.Endpoint, owner)
		}
		var steps []middleware.Step
		for _, name := range api.Middlewares {
			// Permissions authenticate on their own.
			if len(api.Permissions) > 0 && name == "authenticate" {
				continue
			}
			step, err := middleware.Resolve(name, m.window, m.rateLimit)
			if err != nil {
				return err
			}
			steps = append(steps, step)
		}
		if len(api.Permissions) > 0 {
			steps = append(steps, middleware.Permit(api.Permissions)...)
		}
		if api.Validation != nil {
			steps = append(steps, middleware.Validate(api.Validation))
		}
		opened = append(opened, &Open{Owner: owner, Index: index, Meta: api, segments: strings.Split(api.Endpoint, "/"), steps: steps, invoke: invoke})
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.open[owner]; !ok {
		m.order = append(m.order, owner)
	}
	m.open[owner] = opened
	return nil
}

// Remove closes a quan's apis.
func (m *Manager) Remove(owner string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.open, owner)
	for i, o := range m.order {
		if o == owner {
			m.order = append(m.order[:i], m.order[i+1:]...)
			break
		}
	}
}

// Find returns the api that listens to method and path, with the path's parameters.
func (m *Manager) Find(method, path string) (*Open, map[string]string) {
	segments := strings.Split("/"+strings.Trim(path, "/"), "/")
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, owner := range m.order {
		for _, api := range m.open[owner] {
			if api.Meta.Method != method || len(api.segments) != len(segments) {
				continue
			}
			params := map[string]string{}
			matched := true
			for i, s := range api.segments {
				if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
					params[s[1:len(s)-1]] = segments[i]
				} else if s != segments[i] {
					matched = false
					break
				}
			}
			if matched {
				return api, params
			}
		}
	}
	return nil, nil
}

// Handle runs an api's checks, then hands the request to its quan.
func (m *Manager) Handle(ctx context.Context, api *Open, c *middleware.Context) (wire.Response, error) {
	for _, step := range api.steps {
		if err := step(c); err != nil {
			return wire.Response{}, err
		}
	}
	body, err := json.Marshal(c.Body)
	if err != nil {
		return wire.Response{}, apierr.Unexpected()
	}
	return api.invoke(ctx, api.Index, wire.Request{
		RequestID: c.RequestID, Method: c.Method, Path: c.Path, Params: c.Params, Query: c.Query,
		Headers: c.Headers, Body: body, User: c.User,
	})
}

// Described is an api as /quans shows it.
type Described struct {
	Method      string   `json:"method"`
	Endpoint    string   `json:"endpoint"`
	Permissions []string `json:"permissions"`
	Middlewares []string `json:"middlewares"`
	Validated   bool     `json:"validated"`
	Description string   `json:"description"`
}

// Describe lists every quan's apis.
func (m *Manager) Describe() map[string][]Described {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := map[string][]Described{}
	for owner, apis := range m.open {
		for _, a := range apis {
			out[owner] = append(out[owner], Described{a.Meta.Method, a.Meta.Endpoint, a.Meta.Permissions, a.Meta.Middlewares, a.Meta.Validation != nil, a.Meta.Description})
		}
	}
	return out
}
