package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/thrgamon/project-template/internal/auth"
	"github.com/thrgamon/project-template/internal/config"
	"github.com/thrgamon/project-template/internal/domain"
)

type HandlerConfig struct {
	Auth *auth.Service
	Cfg  config.Config
}

type Handler struct {
	auth *auth.Service
	cfg  config.Config
}

func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{auth: cfg.Auth, cfg: cfg.Cfg}
}

// Routes registers all HTTP routes on the given router group.
func (h *Handler) Routes(rg *gin.RouterGroup) {
	rg.GET("/health", h.Health)

	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/logout", h.Logout)
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

// Register creates a user and starts a session.
func (h *Handler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	resp, token, err := h.auth.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	h.setSessionCookie(c, token)
	c.JSON(http.StatusCreated, resp)
}

// Login authenticates a user and starts a session.
func (h *Handler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	resp, token, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: err.Error()})
		return
	}

	h.setSessionCookie(c, token)
	c.JSON(http.StatusOK, resp)
}

// Logout ends the current session.
func (h *Handler) Logout(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err == nil && token != "" {
		_ = h.auth.Logout(c.Request.Context(), token)
	}

	h.clearSessionCookie(c)
	c.JSON(http.StatusOK, domain.MessageResponse{Message: "logged out"})
}

// Me reports the user behind the current session.
func (h *Handler) Me(c *gin.Context) {
	user, ok := auth.GetUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "not authenticated"})
		return
	}

	c.JSON(http.StatusOK, domain.AuthResponse{User: user})
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

func (h *Handler) setSessionCookie(c *gin.Context, token string) {
	maxAge := int(h.cfg.SessionMaxAge.Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("session_token", token, maxAge, "/", h.cfg.CookieDomain, h.cfg.CookieSecure, true)
}

func (h *Handler) clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("session_token", "", -1, "/", h.cfg.CookieDomain, h.cfg.CookieSecure, true)
}
