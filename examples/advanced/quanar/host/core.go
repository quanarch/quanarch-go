package host

import (
	"context"
	"encoding/json"

	shopqua "shop-qua"
)

// core is shop-qua's QuanarCore, as the host implements it: every call goes to the quanar.
type core struct{}

func (core) Db() shopqua.DbCapability         { return db{} }
func (core) Auth() shopqua.AuthCapability     { return auth{} }
func (core) Events() shopqua.EventsCapability { return events{} }
func (core) Log() shopqua.Logger              { return logger{} }

type db struct{}

func (db) Run(ctx context.Context, op shopqua.DbOperation) (json.RawMessage, error) {
	var out json.RawMessage
	if err := call(ctx, "shop-qua:db.run", &out, op); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		out = json.RawMessage("null") // no row
	}
	return out, nil
}

func (db) Transaction(ctx context.Context, ops []shopqua.DbOperation) ([]json.RawMessage, error) {
	var out []json.RawMessage
	return out, call(ctx, "shop-qua:db.transaction", &out, ops)
}

type auth struct{}

func (auth) IssueToken(ctx context.Context, claims shopqua.TokenClaims) (shopqua.IssuedToken, error) {
	var out shopqua.IssuedToken
	return out, call(ctx, "shop-qua:auth.issue_token", &out, claims)
}

func (auth) HashPassword(ctx context.Context, password string) (string, error) {
	var out string
	return out, call(ctx, "shop-qua:auth.hash_password", &out, password)
}

func (auth) VerifyPassword(ctx context.Context, password, hashed string) (bool, error) {
	var out bool
	return out, call(ctx, "shop-qua:auth.verify_password", &out, password, hashed)
}

type events struct{}

func (events) Emit(ctx context.Context, event string, payload any) error {
	return call(ctx, "shop-qua:events.emit", nil, event, payload)
}

// logger sends the quan's log lines; the quanar writes them under the quan's name.
type logger struct{}

func (logger) Info(message string, data any)  { send(logLine("info", message, data)) }
func (logger) Warn(message string, data any)  { send(logLine("warn", message, data)) }
func (logger) Error(message string, data any) { send(logLine("error", message, data)) }

func logLine(level, text string, data any) message {
	return message{Type: "log", Level: level, Message: text, Data: data}
}
