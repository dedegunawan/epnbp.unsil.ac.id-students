package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderXRequestID = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader(HeaderXRequestID) == "" {
			c.Request.Header.Set(HeaderXRequestID, uuid.NewString())
		}
		c.Writer.Header().Set(HeaderXRequestID, c.GetHeader(HeaderXRequestID))
		c.Next()
	}
}
