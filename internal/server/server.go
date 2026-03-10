package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
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

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowMethods = []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Content-Type", "Authorization"}
	corsConfig.AllowCredentials = true
	if opts.Config.Environment == "production" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	}
	engine.Use(cors.New(corsConfig))
	engine.Use(middleware.Logger())

	registerRoutes(engine, opts)

	return &Server{engine: engine}
}

func (s *Server) Run(addr string) error {
	s.http = &http.Server{
		Addr:    addr,
		Handler: s.engine,
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
