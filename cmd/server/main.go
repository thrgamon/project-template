package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/thrgamon/project-template/internal/api"
	"github.com/thrgamon/project-template/internal/auth"
	"github.com/thrgamon/project-template/internal/config"
	"github.com/thrgamon/project-template/internal/db"
	"github.com/thrgamon/project-template/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
	handler := api.NewHandler(authSvc, cfg)

	srv := server.New(server.Options{
		Config:  cfg,
		Handler: handler,
		Auth:    authSvc,
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
