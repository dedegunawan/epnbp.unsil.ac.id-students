package middleware

import (
	"bytes"
	"net/http"
	"runtime/debug"

	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/response"
	"github.com/gin-gonic/gin"
)

func Recovery(lg *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				lg.Errorw("panic", "error", rec)

				stack := debug.Stack()

				// Log error panic dengan stack trace
				lg.Info("stack trace:\n%s", string(bytes.TrimSpace(stack)))

				response.Error(c, http.StatusInternalServerError, "internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
