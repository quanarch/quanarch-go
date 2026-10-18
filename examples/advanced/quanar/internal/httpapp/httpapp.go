// Package httpapp is the quanar's HTTP server: its own routes, then whatever the quans asked it
// to listen to.
package httpapp

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"shop-quanar/internal/apierr"
	"shop-quanar/internal/apis"
	"shop-quanar/internal/auth"
	"shop-quanar/internal/config"
	"shop-quanar/internal/middleware"
	"shop-quanar/internal/wire"
)

// privateHeaders stay with the quanar: a quan never sees the token.
var privateHeaders = map[string]bool{"Authorization": true, "Cookie": true}

func requestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, e *apierr.Error, id string) {
	body := map[string]any{"code": e.Code, "message": e.Message}
	if e.Details != nil {
		body["details"] = e.Details
	}
	writeJSON(w, e.Status, map[string]any{"error": body, "request_id": id})
}

// App answers the quans' apis.
type App struct {
	settings config.Settings
	apis     *apis.Manager
	auth     *auth.Manager
}

// New returns the handler: the system routes first, then the quans.
func New(settings config.Settings, a *apis.Manager, au *auth.Manager, system *http.ServeMux) http.Handler {
	app := &App{settings, a, au}
	system.HandleFunc("/", app.dispatch)
	return system
}

func (a *App) dispatch(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	id := r.Header.Get("X-Request-Id")
	if id == "" {
		id = requestID()
	}
	w.Header().Set("X-Request-Id", id)
	var user *wire.User
	if token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		user = a.auth.Verify(strings.TrimSpace(token))
	}
	status := a.serve(w, r, id, user)
	who := ""
	if user != nil && len(user.ID) >= 8 {
		who = " (" + user.ID[:8] + ")"
	}
	slog.Info(fmt.Sprintf("%s %s → %d %dms%s", r.Method, r.URL.Path, status, time.Since(started).Milliseconds(), who), "scope", "http")
}

func (a *App) serve(w http.ResponseWriter, r *http.Request, id string, user *wire.User) int {
	api, params := a.apis.Find(r.Method, r.URL.Path)
	if api == nil {
		e := apierr.New(http.StatusNotFound, apierr.NotFound, "Nobody listens to "+r.Method+" "+r.URL.Path)
		writeError(w, e, id)
		return e.Status
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, a.settings.BodyLimit+1))
	if err != nil || int64(len(raw)) > a.settings.BodyLimit {
		e := apierr.New(http.StatusRequestEntityTooLarge, apierr.TooLarge, "The body is too large")
		writeError(w, e, id)
		return e.Status
	}
	var body any
	if len(bytes.TrimSpace(raw)) > 0 && json.Unmarshal(raw, &body) != nil {
		e := apierr.New(http.StatusBadRequest, apierr.BadRequest, "The body is not valid JSON")
		writeError(w, e, id)
		return e.Status
	}
	query := map[string]any{}
	for k := range r.URL.Query() {
		query[k] = r.URL.Query().Get(k)
	}
	headers := map[string]string{}
	for k := range r.Header {
		if !privateHeaders[k] {
			headers[strings.ToLower(k)] = r.Header.Get(k)
		}
	}
	client, _, _ := net.SplitHostPort(r.RemoteAddr)
	c := &middleware.Context{RequestID: id, User: user, Client: client, Method: r.Method, Path: r.URL.Path, Params: params, Query: query, Headers: headers, Body: body}
	res, err := a.apis.Handle(r.Context(), api, c)
	if err != nil {
		var e *apierr.Error
		if !errors.As(err, &e) {
			if !errors.Is(err, context.Canceled) {
				slog.Error(r.Method+" "+r.URL.Path+" failed", "error", err, "scope", "http")
			}
			e = apierr.Unexpected()
		}
		writeError(w, e, id)
		return e.Status
	}
	for k, v := range res.Headers {
		w.Header().Set(k, v)
	}
	status := res.Status
	if status == 0 {
		status = http.StatusOK
	}
	if len(res.Body) == 0 {
		w.WriteHeader(status)
		return status
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(res.Body)
	return status
}
