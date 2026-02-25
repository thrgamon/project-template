package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAuth(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		session, err := svc.ValidateSession(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			c.Abort()
			return
		}

		c.Set("user_id", session.UserID)
		c.Set("user_email", session.UserEmail)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (int32, bool) {
	id, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	userID, ok := id.(int32)
	return userID, ok
}
