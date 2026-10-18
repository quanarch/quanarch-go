// Package quacore is the quanar's side of shop-qua, bound to one quan: its database calls reach
// only its tables, its events carry its name. A quan's process asks for a capability by name;
// only the quas plugged into that quan can be reached.
package quacore

import (
	"context"
	"encoding/json"
	"fmt"

	"shop-quanar/internal/auth"
	"shop-quanar/internal/dbmanager"
	"shop-quanar/internal/events"
)

// Capability does one call; its arguments arrive as JSON.
type Capability func(ctx context.Context, args []json.RawMessage) (any, error)

// Core is what the capabilities work with.
type Core struct {
	Db     *dbmanager.Manager
	Auth   *auth.Manager
	Events *events.Manager
}

func arg[T any](args []json.RawMessage, i int) (T, error) {
	var v T
	if i >= len(args) {
		return v, fmt.Errorf("missing argument %d", i)
	}
	return v, json.Unmarshal(args[i], &v)
}

func (c Core) shopQua(owner string) map[string]Capability {
	return map[string]Capability{
		"db.run": func(ctx context.Context, a []json.RawMessage) (any, error) {
			op, err := arg[dbmanager.Operation](a, 0)
			if err != nil {
				return nil, err
			}
			return c.Db.Run(ctx, owner, op)
		},
		"db.transaction": func(ctx context.Context, a []json.RawMessage) (any, error) {
			ops, err := arg[[]dbmanager.Operation](a, 0)
			if err != nil {
				return nil, err
			}
			return c.Db.Transaction(ctx, owner, ops)
		},
		"auth.issue_token": func(_ context.Context, a []json.RawMessage) (any, error) {
			claims, err := arg[auth.Claims](a, 0)
			if err != nil {
				return nil, err
			}
			return c.Auth.Issue(claims), nil
		},
		"auth.hash_password": func(_ context.Context, a []json.RawMessage) (any, error) {
			password, err := arg[string](a, 0)
			if err != nil {
				return nil, err
			}
			return auth.HashPassword(password)
		},
		"auth.verify_password": func(_ context.Context, a []json.RawMessage) (any, error) {
			password, err := arg[string](a, 0)
			if err != nil {
				return nil, err
			}
			hashed, err := arg[string](a, 1)
			if err != nil {
				return nil, err
			}
			return auth.VerifyPassword(password, hashed), nil
		},
		"events.emit": func(_ context.Context, a []json.RawMessage) (any, error) {
			event, err := arg[string](a, 0)
			if err != nil {
				return nil, err
			}
			if len(a) < 2 {
				return nil, fmt.Errorf("missing payload")
			}
			_, err = c.Events.Publish(owner, event, a[1])
			return nil, err
		},
	}
}

// Plugged returns the capabilities of the quas plugged into a quan, by "<qua>:<name>".
func (c Core) Plugged(owner string, quas []string) map[string]Capability {
	cores := map[string]func(string) map[string]Capability{
		"shop-qua": c.shopQua,
	}
	out := map[string]Capability{}
	for _, qua := range quas {
		if core, ok := cores[qua]; ok {
			for name, fn := range core(owner) {
				out[qua+":"+name] = fn
			}
		}
	}
	return out
}
