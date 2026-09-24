package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	sharedauth "github.com/thrgamon/infra/go/auth"

	"github.com/thrgamon/project-template/internal/domain"
)

// contextUserKey is the gin context key under which the authenticated user is
// stored. Handlers should read it via GetUser rather than by name.
const contextPrincipalKey = "auth.principal"

// RequireAuth adapts the shared net/http authenticator to Gin and stores the
// mapped local identity for handlers.
func RequireAuth(app *sharedauth.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, err := app.Authenticate(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "authentication required"})
			return
		}
		c.Set(contextPrincipalKey, principal)
		c.Next()
	}
}

func RequireCSRF(app *sharedauth.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := app.CheckCSRF(c.Request); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, domain.ErrorResponse{Error: "csrf validation failed"})
			return
		}
		c.Next()
	}
}

// GetUser converts the explicit identity mapping's local integer ID to the
// template's API user. Invalid mappings are denied rather than defaulted.
func GetUser(c *gin.Context) (domain.UserResponse, bool) {
	v, exists := c.Get(contextPrincipalKey)
	if !exists {
		return domain.UserResponse{}, false
	}
	p, ok := v.(sharedauth.Principal)
	if !ok || p.Status != "active" {
		return domain.UserResponse{}, false
	}
	id, err := strconv.ParseInt(p.LocalUserID, 10, 32)
	if err != nil || id <= 0 {
		return domain.UserResponse{}, false
	}
	return domain.UserResponse{ID: int32(id), Email: p.Email}, true
}
