package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/thrgamon/project-template/internal/domain"
)

// contextUserKey is the gin context key under which the authenticated user is
// stored. Handlers should read it via GetUser rather than by name.
const contextUserKey = "auth.user"

// RequireAuth rejects requests without a valid session cookie and attaches the
// authenticated user to the request context.
func RequireAuth(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "authentication required"})
			return
		}

		session, err := svc.ValidateSession(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "invalid session"})
			return
		}

		c.Set(contextUserKey, domain.UserResponse{
			ID:    session.UserID,
			Email: session.UserEmail,
		})
		c.Next()
	}
}

// GetUser returns the user attached by RequireAuth. The second return value is
// false on any unauthenticated request, so callers never see a partially
// populated user.
func GetUser(c *gin.Context) (domain.UserResponse, bool) {
	v, exists := c.Get(contextUserKey)
	if !exists {
		return domain.UserResponse{}, false
	}
	user, ok := v.(domain.UserResponse)
	return user, ok
}
