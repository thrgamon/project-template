package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// contextRequestIDKey is the gin context key holding the per-request ID.
// Read it via GetRequestID rather than by name.
const contextRequestIDKey = "middleware.request_id"

// RequestID assigns each request a UUID, echoes it as X-Request-ID and stores
// it on the context for the logger.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.New().String()
		c.Set(contextRequestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// GetRequestID returns the ID assigned by RequestID, or false if the
// middleware did not run for this request.
func GetRequestID(c *gin.Context) (string, bool) {
	v, exists := c.Get(contextRequestIDKey)
	if !exists {
		return "", false
	}
	id, ok := v.(string)
	return id, ok
}
