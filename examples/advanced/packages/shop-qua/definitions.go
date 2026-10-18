package shopqua

import (
	"context"
	"encoding/json"
)

// What a quan declares. The quanar's managers turn each declaration into something real: a
// table, a route, a check, a subscription. A quan writes none of those by hand.

// FieldDefinition is one column of a model.
type FieldDefinition struct {
	Type     FieldType `json:"type"`
	Required bool      `json:"required"`
	Unique   bool      `json:"unique"`
	Index    bool      `json:"index"`
	Default  any       `json:"default"`
	// Values are an enum's values.
	Values []string `json:"values"`
}

// ModelDefinition is a table the quan owns. The quanar creates and migrates it; id, created_at
// and updated_at are added to every model.
type ModelDefinition struct {
	Name   string                     `json:"name"`
	Fields map[string]FieldDefinition `json:"fields"`
	// Indexes are indexes over several fields.
	Indexes [][]string `json:"indexes"`
}

// PermissionDefinition is a permission a quan declares. Who holds it is not the quan's or the
// quanar's business.
type PermissionDefinition struct {
	// Key starts with the quan's name: products:manage
	Key         string `json:"key"`
	Description string `json:"description"`
}

// Rule is one value's rule. The quanar checks it before the request reaches the quan.
type Rule struct {
	Type     RuleType `json:"type"`
	Required bool     `json:"required"`
	// Min and Max are a length for text and lists, a value for numbers.
	Min *float64 `json:"min"`
	Max *float64 `json:"max"`
	// Pattern is a regular expression text must match.
	Pattern string   `json:"pattern"`
	Values  []string `json:"values"`
	// Items is the rule of each item of a list.
	Items *Rule `json:"items"`
	// Fields are the rules of an object's fields.
	Fields map[string]Rule `json:"fields"`
}

// ValidationRules are the rules of a request's path, query and body.
type ValidationRules struct {
	Params map[string]Rule `json:"params"`
	Query  map[string]Rule `json:"query"`
	Body   map[string]Rule `json:"body"`
}

// Handler answers a request.
type Handler func(ctx context.Context, req *ApiRequest) (ApiResponse, error)

// Next runs what comes after a middleware: the next middleware, or the handler.
type Next func() (ApiResponse, error)

// LocalMiddleware is a middleware of the quan itself. It runs in the quan, after the quanar's checks.
type LocalMiddleware func(ctx context.Context, req *ApiRequest, next Next) (ApiResponse, error)

// ApiDefinition is an api the quan asks the quanar to listen to. The quanar opens the route,
// runs its middlewares, checks its permissions and validation, and only then calls the quan.
type ApiDefinition struct {
	// Endpoint starts with the quan's name: /products, /products/{id}
	Endpoint string
	Method   HttpMethod
	Handler  Handler
	// QuanarMiddlewares run in the quanar, first.
	QuanarMiddlewares []QuanarMiddleware
	// Middlewares run in the quan, after the quanar's checks.
	Middlewares []LocalMiddleware
	// Permissions are all required. They imply authentication.
	Permissions []string
	Validation  *ValidationRules
	Description string
}

// EventHandler handles an event another quan announced.
type EventHandler func(ctx context.Context, payload json.RawMessage, meta EventMeta) error

// EventSubscription is an event a quan listens to. Emitters and listeners never know each
// other's code.
type EventSubscription struct {
	// Event is <quan>.<name>, for example orders.placed
	Event   string
	Handler EventHandler
}

// QuanManifest is everything a quan is, declared. The quanar's quan registry turns it into a
// running quan.
type QuanManifest struct {
	Name        string
	Version     string
	Models      []ModelDefinition
	Permissions []PermissionDefinition
	Apis        []ApiDefinition
	Events      []EventSubscription
	// OnMount runs once the quanar has created the models and opened the apis.
	OnMount   func(ctx context.Context) error
	OnUnmount func(ctx context.Context) error
}

// Num is a number for a Rule's Min and Max.
func Num(n float64) *float64 { return &n }
