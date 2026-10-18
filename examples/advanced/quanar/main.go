// The shop's quanar: one server process. It opens the apis, creates the tables, checks
// permissions and validation and carries events for the quans it loads, each in its own process.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shop-quanar/gate"
	"shop-quanar/internal/apis"
	"shop-quanar/internal/auth"
	"shop-quanar/internal/config"
	"shop-quanar/internal/database"
	"shop-quanar/internal/dbmanager"
	"shop-quanar/internal/events"
	"shop-quanar/internal/httpapp"
	"shop-quanar/internal/permissions"
	"shop-quanar/internal/quacore"
	"shop-quanar/internal/quans"
	"shop-quanar/internal/registry"
	"shop-quanar/internal/schema"
	"shop-quanar/internal/system"
)

func main() {
	// Started as a quan's gate: close the walls and become the quan.
	if gate.Main() {
		return
	}
	settings := config.Load()
	if !gate.Walled {
		slog.Warn("quans run without walls on this operating system: no file, network or memory limits")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(settings.DatabaseFile)
	if err != nil {
		slog.Error("database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	schemas := schema.New(db)
	perms := permissions.New()
	apiManager := apis.New(settings.RateLimitWindow, settings.RateLimitMax)
	eventManager := events.New()
	authManager := auth.New(settings.JWTSecret, settings.TokenTTL)
	core := quacore.Core{Db: dbmanager.New(db, schemas), Auth: authManager, Events: eventManager}
	service := quans.New(settings, registry.New(settings, core, schemas, perms, apiManager, eventManager))

	handler := httpapp.New(settings, apiManager, authManager, system.Routes(service, apiManager, schemas, perms, eventManager))
	server := &http.Server{Addr: fmt.Sprintf(":%d", settings.Port), Handler: handler}
	go func() {
		slog.Info(fmt.Sprintf("listening on http://localhost:%d", settings.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server", "error", err)
			stop()
		}
	}()
	service.Start(ctx)

	<-ctx.Done()
	slog.Info("stopping")
	service.Stop()
	shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
}
