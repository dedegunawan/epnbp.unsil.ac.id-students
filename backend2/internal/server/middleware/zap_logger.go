package middleware

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
	"github.com/gin-gonic/gin"
	"time"
)

func ZapLogger(lg *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		lat := time.Since(start)
		lg.Infow("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"ip", c.ClientIP(),
			"latency_ms", lat.Milliseconds(),
		)
	}
}
