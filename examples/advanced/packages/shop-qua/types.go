package shopqua

import (
	"context"
	"encoding/json"
)

// What crosses between a quan and the quanar while it runs, and the core the quanar plugs in.

// AuthUser is who is calling, as the quanar read it from the request's token.
type AuthUser struct {
	// ID is the subject the token was issued for, usually a user id.
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// TokenClaims are what a quan asks the quanar to put in a token.
type TokenClaims struct {
	Sub         string   `json:"sub"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// IssuedToken is a signed token and when it expires, as ISO time.
type IssuedToken struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// ApiRequest is a request as the quan sees it: already authenticated, permitted and validated.
type ApiRequest struct {
	RequestID string            `json:"request_id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Params    map[string]string `json:"params"`
	Query     map[string]any    `json:"query"`
	Headers   map[string]string `json:"headers"`
	Body      json.RawMessage   `json:"body"`
	User      *AuthUser         `json:"user"`
	// Locals are filled by the quan's own middlewares for the handler that follows.
	Locals map[string]any `json:"-"`
}

// ApiResponse is what a handler answers. A zero Status means 200.
type ApiResponse struct {
	Status  int               `json:"status,omitempty"`
	Body    any               `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// EventMeta says who emitted an event and when.
type EventMeta struct {
	EmittedBy string `json:"emitted_by"`
	At        string `json:"at"`
}

// Where is field → value, or field → map of "eq", "ne", "gt", "gte", "lt", "lte", "in", "like".
type Where map[string]any

// DbOperation is one database operation, as data: it crosses from the quan to the quanar.
type DbOperation struct {
	Action  string            `json:"action"` // find, find_one, count, create, update, delete
	Model   string            `json:"model"`
	Where   Where             `json:"where,omitempty"`
	OrderBy map[string]string `json:"order_by,omitempty"`
	Limit   int               `json:"limit,omitempty"`
	Offset  int               `json:"offset,omitempty"`
	// Data is for create and update. A number may change in place: {"stock": {"decrement": 2}}.
	Data map[string]any `json:"data,omitempty"`
	// Expect, for update and delete: any other number of rows fails the operation, and its transaction.
	Expect *int `json:"expect,omitempty"`
}

// DbCapability is the quanar's database, seen through a quan: only that quan's models exist.
type DbCapability interface {
	Run(ctx context.Context, op DbOperation) (json.RawMessage, error)
	// Transaction runs every operation or none.
	Transaction(ctx context.Context, ops []DbOperation) ([]json.RawMessage, error)
}

// AuthCapability is identity work the quanar does: signing and checking secrets never leave it.
type AuthCapability interface {
	IssueToken(ctx context.Context, claims TokenClaims) (IssuedToken, error)
	HashPassword(ctx context.Context, password string) (string, error)
	VerifyPassword(ctx context.Context, password, hashed string) (bool, error)
}

// EventsCapability carries what a quan announces. The name starts with the quan's name.
type EventsCapability interface {
	Emit(ctx context.Context, event string, payload any) error
}

// Logger writes the quan's log lines under its name.
type Logger interface {
	Info(message string, data any)
	Warn(message string, data any)
	Error(message string, data any)
}

// QuanarCore is what the quanar implements and plugs in with Init.
type QuanarCore interface {
	Db() DbCapability
	Auth() AuthCapability
	Events() EventsCapability
	Log() Logger
}
