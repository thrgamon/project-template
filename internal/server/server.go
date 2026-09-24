package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/thrgamon/project-template/internal/api"
	"github.com/thrgamon/project-template/internal/config"
	"github.com/thrgamon/project-template/internal/middleware"
)

type Options struct {
	Config  config.Config
	Handler *api.Handler
}

type Server struct {
	engine *gin.Engine
	http   *http.Server
}

var ErrServerClosed = errors.New("server closed")

func New(opts Options) *Server {
	if opts.Config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()
	engine.Use(otelgin.Middleware("myapp"))
	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestID())

	// The SvelteKit frontend calls /api on the same origin. Keeping browser API
	// requests same-origin lets host-only session cookies and CSRF origin checks
	// apply without a permissive credentialed CORS policy.
	engine.Use(middleware.Logger())

	registerRoutes(engine, opts)
	registerStaticFiles(engine, opts.Config.StaticDir)

	return &Server{engine: engine}
}

func (s *Server) Run(addr string) error {
	s.http = &http.Server{
		Addr:    addr,
		Handler: s.engine,
		// Bound how long a slow or idle peer can hold a connection open.
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	fmt.Printf("HTTP server listening on %s\n", addr)

	if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return ErrServerClosed
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}

func registerRoutes(router *gin.Engine, opts Options) {
	apiGroup := router.Group("/api")
	opts.Handler.Routes(apiGroup)
}

// registerStaticFiles serves the SvelteKit adapter-static build from the same
// origin as the API. Local development leaves StaticDir empty and Vite proxies
// /api requests to the Go process instead.
func registerStaticFiles(router *gin.Engine, staticDir string) {
	if staticDir == "" {
		return
	}
	if _, err := os.Stat(staticDir); err != nil {
		panic(fmt.Sprintf("static frontend directory %q is unavailable: %v", staticDir, err))
	}

	files := http.FileServer(http.Dir(staticDir))
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Status(http.StatusNotFound)
			return
		}
		files.ServeHTTP(c.Writer, c.Request)
	})
}
