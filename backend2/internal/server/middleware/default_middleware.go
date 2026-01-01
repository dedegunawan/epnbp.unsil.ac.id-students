package middleware

import "github.com/gin-gonic/gin"

func DefaultMiddleware(...any) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
