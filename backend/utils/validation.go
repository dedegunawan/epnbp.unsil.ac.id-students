package utils

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"
)

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
