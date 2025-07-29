package routes

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r gin.IRoutes) {
	r.GET("/sso-login", controllers.SsoLoginHandler)
	r.GET("/sso-logout", controllers.SsoLogoutHandler)
	r.POST("/login", controllers.LoginHandler)
	r.GET("/callback", controllers.CallbackHandler)
}
