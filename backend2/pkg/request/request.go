package request

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"strings"
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

func RespondValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if ok := strings.Contains(err.Error(), "validation for"); err != nil && ok {
		if errors.As(err, &ve) {
			out := make(map[string]string)
			for _, fe := range ve {
				field := strings.ToLower(fe.Field())
				switch fe.Tag() {
				case "required":
					out[field] = field + " is required"
				case "email":
					out[field] = "invalid email format"
				case "min":
					out[field] = field + " must be at least " + fe.Param() + " characters"
				default:
					out[field] = "invalid value"
				}
			}
			c.JSON(http.StatusBadRequest, gin.H{"errors": out})
			return
		}
	}

	// fallback if not validation error
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
