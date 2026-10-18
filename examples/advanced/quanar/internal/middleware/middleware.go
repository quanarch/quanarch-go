// Package middleware holds the quanar's middlewares. Quans name them; this is where a name
// becomes code, so a quan can never bring its own authentication or skip a check.
package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"shop-quanar/internal/apierr"
	"shop-quanar/internal/permissions"
	"shop-quanar/internal/validation"
	"shop-quanar/internal/wire"
)

// Context is what the quanar knows about a request on its way to a quan.
type Context struct {
	RequestID string
	User      *wire.User
	Client    string
	Method    string
	Path      string
	Params    map[string]string
	Query     map[string]any
	Headers   map[string]string
	// Body is the decoded JSON body; validation coerces it in place.
	Body any
}

// Step is one check before the quan: it passes, or returns the error the client gets.
type Step func(*Context) *apierr.Error

// Authenticate answers 401 unless the request carries a valid token.
func Authenticate(c *Context) *apierr.Error {
	if c.User == nil {
		return apierr.New(http.StatusUnauthorized, apierr.Unauthenticated, "Sign in first")
	}
	return nil
}

// rateLimit is a fixed window per client and api.
func rateLimit(window time.Duration, max int) Step {
	var mu sync.Mutex
	windows := map[string]struct {
		start time.Time
		count int
	}{}
	return func(c *Context) *apierr.Error {
		mu.Lock()
		defer mu.Unlock()
		now := time.Now()
		w, ok := windows[c.Client]
		if !ok || now.Sub(w.start) > window {
			w.start, w.count = now, 0
		}
		w.count++
		windows[c.Client] = w
		if w.count > max {
			return apierr.New(http.StatusTooManyRequests, apierr.RateLimited, "Too many requests, try again later")
		}
		return nil
	}
}

// Resolve turns a middleware's name into the quanar's code.
func Resolve(name string, window time.Duration, max int) (Step, error) {
	switch name {
	case "authenticate":
		return Authenticate, nil
	case "rate-limit":
		return rateLimit(window, max), nil
	}
	return nil, fmt.Errorf("the quanar has no middleware %q", name)
}

// Permit authenticates, then checks every required permission.
func Permit(required []string) []Step {
	return []Step{Authenticate, func(c *Context) *apierr.Error {
		if missing := permissions.Missing(c.User, required); len(missing) > 0 {
			return apierr.New(http.StatusForbidden, apierr.Forbidden, "Needs "+strings.Join(missing, ", ")).With(map[string]any{"missing": missing})
		}
		return nil
	}}
}

// Validate checks the request against an api's rules, coercing path and query values.
func Validate(rules *wire.Validation) Step {
	return func(c *Context) *apierr.Error {
		params := map[string]any{}
		for k, v := range c.Params {
			params[k] = v
		}
		issues := validation.Validate(rules, params, c.Query, c.Body)
		if len(issues) > 0 {
			return apierr.New(http.StatusUnprocessableEntity, apierr.Validation, "The request is not valid").With(map[string]any{"issues": issues})
		}
		return nil
	}
}
