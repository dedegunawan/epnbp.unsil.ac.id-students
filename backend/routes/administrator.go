package routes

import (
	manage_users "github.com/dedegunawan/backend-ujian-telp-v5/controllers/manage-users"
	"github.com/dedegunawan/backend-ujian-telp-v5/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterAdministrator(r *gin.RouterGroup) {
	RegisterUserRoutes(r)

}

func RegisterUserRoutes(r *gin.RouterGroup) {
	user := r.Group("/users")
	user.Use(middleware.RequireAuthFromTokenDB())
	{
		user.GET("", manage_users.Index)
		user.POST("", manage_users.Create)
		user.PUT(":id", manage_users.Edit)
		user.DELETE(":id", manage_users.Delete)
		user.GET("/export", manage_users.Export)
	}
}
