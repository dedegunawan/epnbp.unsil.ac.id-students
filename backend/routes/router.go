package routes

import (
	"net/http"

	"github.com/dedegunawan/backend-ujian-telp-v5/auth"
	"github.com/dedegunawan/backend-ujian-telp-v5/controllers"
	"github.com/dedegunawan/backend-ujian-telp-v5/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.LoadCors())

	r.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/sso-login")
	})

	auth.InitOIDC()

	RegisterAuthRoutes(r)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/me", middleware.RequireAuthFromTokenDB(), controllers.Me)
		v1.GET("/student-bill", middleware.RequireAuthFromTokenDB(), controllers.GetStudentBillStatus)
		v1.GET("/student-bill-new", middleware.RequireAuthFromTokenDB(), controllers.GetStudentBillStatusNew)
		v1.POST("/student-bill", middleware.RequireAuthFromTokenDB(), controllers.GenerateCurrentBill)
		v1.POST("/regenerate-student-bill", middleware.RequireAuthFromTokenDB(), controllers.RegenerateCurrentBill)
		v1.GET("/generate/:StudentBillID", middleware.RequireAuthFromTokenDB(), controllers.GenerateUrlPembayaran)
		v1.GET("/generate-payment-new", middleware.RequireAuthFromTokenDB(), controllers.GenerateUrlPembayaranNew)
		v1.POST("/confirm-payment/:StudentBillID", middleware.RequireAuthFromTokenDB(), controllers.ConfirmPembayaran)
		v1.GET("/back-to-sintesys", middleware.RequireAuthFromTokenDB(), controllers.BackToSintesys)

		// Payment status endpoints
		v1.GET("/payment-status", middleware.RequireAuthFromTokenDB(), controllers.GetPaymentStatus)
		v1.GET("/payment-status/summary", middleware.RequireAuthFromTokenDB(), controllers.GetPaymentStatusSummary)
		v1.PUT("/payment-status/:id", middleware.RequireAuthFromTokenDB(), controllers.UpdatePaymentStatus)

		// Student bills endpoints (public, no auth required)
		v1.GET("/student-bills", controllers.GetAllStudentBills)

		// Payment status logs endpoints (public, no auth required)
		v1.GET("/payment-status-logs", controllers.GetAllPaymentStatusLogs)

		// Worker endpoints dihapus - tidak ada operasi write ke database

		v1.GET("/payment-callback", controllers.PaymentCallbackHandler)
		v1.POST("/payment-callback", controllers.PaymentCallbackHandler)
	}
	RegisterAdministrator(v1)

	return r
}
