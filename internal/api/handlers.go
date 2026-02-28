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

// Health godoc
// @Summary Health check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body domain.RegisterRequest true "Registration details"
// @Success 201 {object} domain.AuthResponse
// @Failure 400 {object} map[string]string
// @Router /api/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, token, err := h.auth.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.setSessionCookie(c, token)
	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param body body domain.LoginRequest true "Login credentials"
// @Success 200 {object} domain.AuthResponse
// @Failure 401 {object} map[string]string
// @Router /api/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, token, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.setSessionCookie(c, token)
	c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary Logout current session
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err == nil && token != "" {
		_ = h.auth.Logout(c.Request.Context(), token)
	}

	h.clearSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// Me godoc
// @Summary Get current user
// @Tags auth
// @Produce json
// @Success 200 {object} domain.UserResponse
// @Failure 401 {object} map[string]string
// @Router /api/auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	userID, ok := auth.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	email, _ := c.Get("user_email")
	c.JSON(http.StatusOK, domain.UserResponse{
		ID:    userID,
		Email: email.(string),
	})
}

// Dashboard godoc
// @Summary Example protected endpoint
// @Tags dashboard
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/dashboard [get]
func (h *Handler) Dashboard(c *gin.Context) {
	email, _ := c.Get("user_email")
	c.JSON(http.StatusOK, gin.H{
		"message": "welcome to the dashboard",
		"email":   email,
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
