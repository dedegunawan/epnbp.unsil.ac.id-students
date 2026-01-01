package response

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
)

var ERROR_TO_FRONTEND = 1

var ErrorOption = ""

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(200, APIResponse{Success: true, Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(201, APIResponse{Success: true, Data: data})
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(code, APIResponse{Success: false, Error: msg})
}

func ErrorHandler(c *gin.Context, code int, message string) {
	opt := 0
	ErrorOption, err := strconv.Atoi(os.Getenv("ERROR_OPTION"))
	if err == nil {
		opt = ErrorOption
	}

	if opt == ERROR_TO_FRONTEND {
		url := os.Getenv("FRONTEND_ERROR_URL")
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?code=%d&error=%s", url, code, message))
		return
	} else {
		c.JSON(code, gin.H{"error": message})
		return
	}
}
