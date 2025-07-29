package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
)

var ERROR_TO_FRONTEND = 1

var ErrorOption = ""

func ErrorHandler(c *gin.Context, code int, message string) {
	opt := 0
	ErrorOption, err := strconv.Atoi(os.Getenv("ERROR_OPTION"))
	if err == nil {
		opt = ErrorOption
	}

	Log.Info("ERROR_OPTION : ", ErrorOption)

	if opt == ERROR_TO_FRONTEND {
		url := os.Getenv("FRONTEND_ERROR_URL")
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?code%d&message=%s", url, code, message))
		return
	} else {
		c.JSON(code, gin.H{"error": message})
		return
	}
}
