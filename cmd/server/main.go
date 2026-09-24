package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	sharedauth "github.com/thrgamon/infra/go/auth"

	"github.com/thrgamon/project-template/internal/api"
	"github.com/thrgamon/project-template/internal/config"
	"github.com/thrgamon/project-template/internal/server"
	"github.com/thrgamon/project-template/internal/telemetry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Telemetry (no-op if OTEL_EXPORTER_OTLP_ENDPOINT is unset)
	logger, shutdownTelemetry, err := telemetry.Init(ctx, "myapp")
	if err != nil {
		log.Fatalf("initialize telemetry: %v", err)
	}
	defer func() { _ = shutdownTelemetry(context.Background()) }()
	_ = logger // use logger instead of slog.Default() throughout

	cfg := config.LoadConfig()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	if err := pool.Ping(pingCtx); err != nil {
		cancelPing()
		log.Fatalf("ping database: %v", err)
	}
	cancelPing()

	store := sharedauth.NewPGStore(pool)
	authApp, err := sharedauth.New(ctx, sharedauth.Config{
		IssuerURL:            cfg.Auth0IssuerURL,
		ClientID:             cfg.Auth0ClientID,
		ClientSecret:         cfg.Auth0ClientSecret,
		RedirectURL:          cfg.Auth0RedirectURL,
		CookieName:           "session_token",
		CookieSecure:         cfg.CookieSecure,
		AllowInsecureCookies: cfg.AllowInsecureCookies,
		StateSecret:          []byte(cfg.AuthStateSecret),
		SessionMaxAge:        cfg.SessionMaxAge,
	}, store, store)
	if err != nil {
		log.Fatalf("initialize Auth0 authentication: %v", err)
	}
	handler := api.NewHandler(api.HandlerConfig{
		Auth: authApp,
	})

	srv := server.New(server.Options{
		Config:  cfg,
		Handler: handler,
	})

	addr := fmt.Sprintf(":%d", cfg.Port)

	// Background cleanup only deletes expired opaque Auth0 sessions. Identity
	// revocation is immediate because every lookup joins current membership.
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cleanupCtx, cancelCleanup := context.WithTimeout(ctx, 30*time.Second)
				if err := store.DeleteExpiredSessions(cleanupCtx); err != nil {
					slog.Error("cleaning expired sessions", "error", err)
				}
				cancelCleanup()
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		if err := srv.Run(addr); err != nil && !errors.Is(err, server.ErrServerClosed) {
			log.Fatalf("server stopped: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}
