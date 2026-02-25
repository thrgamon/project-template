package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

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

		if reqID, ok := c.Get("request_id"); ok {
			attrs = append(attrs, slog.String("request_id", reqID.(string)))
		}
		if userID, ok := c.Get("user_id"); ok {
			attrs = append(attrs, slog.Any("user_id", userID))
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
