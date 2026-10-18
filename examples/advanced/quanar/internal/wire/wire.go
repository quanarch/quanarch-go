// Package wire holds what travels between the quanar and a quan: a request as the quan sees
// it, and what the quan declared, as JSON. The quanar keeps everything but the quan's code.
package wire

import "encoding/json"

// User is who is calling, as the quanar read it from the token.
type User struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// Request is a request as the quan sees it.
type Request struct {
	RequestID string            `json:"request_id"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Params    map[string]string `json:"params"`
	Query     map[string]any    `json:"query"`
	Headers   map[string]string `json:"headers"`
	Body      json.RawMessage   `json:"body"`
	User      *User             `json:"user"`
}

// Response is what the quan answered.
type Response struct {
	Status  int               `json:"status,omitempty"`
	Body    json.RawMessage   `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Field is one column of a model.
type Field struct {
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Unique   bool     `json:"unique"`
	Index    bool     `json:"index"`
	Default  any      `json:"default"`
	Values   []string `json:"values"`
}

// Model is a table a quan owns.
type Model struct {
	Name    string           `json:"name"`
	Fields  map[string]Field `json:"fields"`
	Indexes [][]string       `json:"indexes"`
}

// Permission is a permission a quan declares.
type Permission struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

// Rule is one value's rule.
type Rule struct {
	Type     string          `json:"type"`
	Required bool            `json:"required"`
	Min      *float64        `json:"min"`
	Max      *float64        `json:"max"`
	Pattern  string          `json:"pattern"`
	Values   []string        `json:"values"`
	Items    *Rule           `json:"items"`
	Fields   map[string]Rule `json:"fields"`
}

// Validation is an api's rules.
type Validation struct {
	Params map[string]Rule `json:"params"`
	Query  map[string]Rule `json:"query"`
	Body   map[string]Rule `json:"body"`
}

// Api is an api as the quanar keeps it.
type Api struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	// Middlewares are the quanar's, by name.
	Middlewares []string    `json:"middlewares"`
	Permissions []string    `json:"permissions"`
	Validation  *Validation `json:"validation"`
	Description string      `json:"description"`
	// LocalMiddlewares is how many of the quan's own middlewares run before its handler.
	LocalMiddlewares int `json:"local_middlewares"`
}

// Manifest is what a quan declared.
type Manifest struct {
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Models      []Model      `json:"models"`
	Permissions []Permission `json:"permissions"`
	Apis        []Api        `json:"apis"`
	// Events are the events it listens to, in the order of its subscriptions.
	Events []string `json:"events"`
}

// Meta says who emitted an event and when.
type Meta struct {
	EmittedBy string `json:"emitted_by"`
	At        string `json:"at"`
}
