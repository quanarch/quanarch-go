// Package system holds the quanar's own routes: its health, and what every quan declared.
package system

import (
	"encoding/json"
	"net/http"
	"time"

	"shop-quanar/internal/apis"
	"shop-quanar/internal/events"
	"shop-quanar/internal/permissions"
	"shop-quanar/internal/quans"
	"shop-quanar/internal/schema"
)

// Routes returns /health and /quans.
func Routes(q *quans.Service, a *apis.Manager, s *schema.Manager, p *permissions.Manager, e *events.Manager) *http.ServeMux {
	started := time.Now()
	mux := http.NewServeMux()
	write := func(w http.ResponseWriter, body any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		write(w, map[string]any{"status": "ok", "uptime": int(time.Since(started).Seconds())})
	})
	mux.HandleFunc("GET /quans", func(w http.ResponseWriter, _ *http.Request) {
		write(w, map[string]any{
			"quans":       q.Statuses(),
			"apis":        a.Describe(),
			"models":      s.Describe(),
			"permissions": p.All(),
			"events":      e.Describe(),
		})
	})
	return mux
}
