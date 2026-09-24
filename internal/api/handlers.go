package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	sharedauth "github.com/thrgamon/infra/go/auth"

	"github.com/thrgamon/project-template/internal/auth"
	"github.com/thrgamon/project-template/internal/domain"
)

type HandlerConfig struct {
	Auth *sharedauth.App
}

type Handler struct {
	auth *sharedauth.App
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{auth: cfg.Auth}
}

// Routes registers all HTTP routes on the given router group.
func (h *Handler) Routes(rg *gin.RouterGroup) {
	rg.GET("/health", h.Health)

	authGroup := rg.Group("/auth")
	{
		authGroup.GET("/login", gin.WrapF(h.auth.Login))
		authGroup.GET("/callback", gin.WrapF(h.auth.Callback))
		authGroup.POST("/logout", gin.WrapF(h.auth.Logout))
		authGroup.GET("/me", auth.RequireAuth(h.auth), h.Me)
	}

	protected := rg.Group("")
	protected.Use(auth.RequireAuth(h.auth))
	{
		protected.GET("/dashboard", h.Dashboard)
	}
}

// Health reports that the process is up and serving.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, domain.HealthResponse{Status: "ok"})
}

// Me reports the user behind the current session.
func (h *Handler) Me(c *gin.Context) {
	user, ok := auth.GetUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "not authenticated"})
		return
	}

	csrfToken, err := h.auth.EnsureCSRF(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "creating csrf token"})
		return
	}
	c.JSON(http.StatusOK, domain.AuthResponse{User: user, CSRFToken: csrfToken})
}

// Dashboard is an example protected endpoint.
func (h *Handler) Dashboard(c *gin.Context) {
	user, ok := auth.GetUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "not authenticated"})
		return
	}

	c.JSON(http.StatusOK, domain.DashboardResponse{
		Message: "welcome to the dashboard",
		Email:   user.Email,
	})
}
