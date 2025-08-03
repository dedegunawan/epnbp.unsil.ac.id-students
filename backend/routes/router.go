package routes

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/auth"
	"github.com/dedegunawan/backend-ujian-telp-v5/controllers"
	"github.com/dedegunawan/backend-ujian-telp-v5/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.LoadCors())

	r.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/sso-login")
		return
	})

	auth.InitOIDC()

	RegisterAuthRoutes(r)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/me", middleware.RequireAuthFromTokenDB(), controllers.Me)
		v1.GET("/student-bill", middleware.RequireAuthFromTokenDB(), controllers.GetStudentBillStatus)
		v1.POST("/student-bill", middleware.RequireAuthFromTokenDB(), controllers.GenerateCurrentBill)
		v1.GET("/generate/:StudentBillID", middleware.RequireAuthFromTokenDB(), controllers.GenerateUrlPembayaran)
		v1.POST("/confirm-payment/:StudentBillID", middleware.RequireAuthFromTokenDB(), controllers.ConfirmPembayaran)
		v1.GET("/back-to-sintesys", middleware.RequireAuthFromTokenDB(), controllers.BackToSintesys)

		v1.GET("/payment-callback", controllers.PaymentCallbackHandler)
		v1.POST("/payment-callback", controllers.PaymentCallbackHandler)
	}
	RegisterAdministrator(v1)

	return r
}
