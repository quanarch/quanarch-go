// Package wire holds a request and a response as they travel between the quanar and a quan.
// The quanar speaks this, never a qua's types: any qua whose shapes match it can be plugged.
package wire

import "encoding/json"

// Request is a request as a quan sees it. The quanar's HTTP server never reaches the quan.
type Request struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Params  map[string]string `json:"params"`
	Query   map[string]string `json:"query"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
}

// Response is what a quan answers. A zero Status means 200.
type Response struct {
	Status int `json:"status,omitempty"`
	Body   any `json:"body,omitempty"`
}
