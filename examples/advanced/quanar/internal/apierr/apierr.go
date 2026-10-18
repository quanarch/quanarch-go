// Package apierr holds the errors the quanar answers with, and their codes.
package apierr

import "net/http"

// Codes the quanar answers with; the client can rely on them.
const (
	NotFound        = "not_found"
	Unauthenticated = "unauthenticated"
	Forbidden       = "forbidden"
	Validation      = "validation_failed"
	RateLimited     = "rate_limited"
	BadRequest      = "bad_request"
	TooLarge        = "too_large"
	QuanUnavailable = "quan_unavailable"
	QuanTimeout     = "quan_timeout"
	QuanFailed      = "quan_failed"
	Internal        = "internal"
)

// Error already knows how it should be answered.
type Error struct {
	Status  int
	Code    string
	Message string
	Details any
}

func (e *Error) Error() string { return e.Message }

// New makes an Error.
func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

// With adds details, for example the issues of a failed validation.
func (e *Error) With(details any) *Error {
	e.Details = details
	return e
}

// Unexpected is what every unexpected failure becomes; its cause is logged, never answered.
func Unexpected() *Error { return New(http.StatusInternalServerError, Internal, "Internal error") }
