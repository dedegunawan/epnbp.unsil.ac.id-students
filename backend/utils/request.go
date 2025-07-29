package utils

import (
	"errors"
	"io"

	"github.com/gin-gonic/gin"
	"net/http"
)

func BindJSONOrAbort(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		if errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request body cannot be empty"})
		} else {
			RespondValidationError(c, err)
		}
		return false
	}
	return true
}
