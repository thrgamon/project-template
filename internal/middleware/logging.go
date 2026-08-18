package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/thrgamon/project-template/internal/auth"
)

// Logger emits one structured line per request. It runs after RequestID and
// (where mounted) RequireAuth, so it can attach both identifiers.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start)

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Duration("duration", duration),
			slog.String("client_ip", c.ClientIP()),
		}

		if reqID, ok := GetRequestID(c); ok {
			attrs = append(attrs, slog.String("request_id", reqID))
		}
		if user, ok := auth.GetUser(c); ok {
			attrs = append(attrs, slog.Int64("user_id", int64(user.ID)))
		}

		msg := "request"
		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		slog.LogAttrs(c.Request.Context(), level, msg, attrs...)
	}
}
