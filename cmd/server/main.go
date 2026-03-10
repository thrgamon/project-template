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

	"github.com/thrgamon/project-template/internal/api"
	"github.com/thrgamon/project-template/internal/auth"
	"github.com/thrgamon/project-template/internal/config"
	"github.com/thrgamon/project-template/internal/db"
	"github.com/thrgamon/project-template/internal/server"
	"github.com/thrgamon/project-template/internal/telemetry"
)

// @title My App API
// @version 1.0
// @host localhost:8080
// @BasePath /api

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

	queries := db.New(pool)
	authSvc := auth.NewService(queries, cfg)
	handler := api.NewHandler(api.HandlerConfig{
		Auth: authSvc,
		Cfg:  cfg,
	})

	srv := server.New(server.Options{
		Config:  cfg,
		Handler: handler,
	})

	addr := fmt.Sprintf(":%d", cfg.Port)

	// Background session cleanup
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := authSvc.DeleteExpiredSessions(context.Background()); err != nil {
					slog.Error("cleaning expired sessions", "error", err)
				}
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
