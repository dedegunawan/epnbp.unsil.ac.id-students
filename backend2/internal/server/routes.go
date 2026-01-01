package server

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/transport/http/auth"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/transport/http/mahasiswa"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/transport/http/user"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	AuthSSO   *auth.AuthSsoHandler
	User      *user.UserHandler
	Mahasiswa *mahasiswa.MahasiswaHandler
	// nanti tambah lagi misalnya Product, Order, dsb.
}
type Middleware struct {
	AuthJWT   gin.HandlerFunc
	CORS      gin.HandlerFunc
	RequestID gin.HandlerFunc
	Logger    gin.HandlerFunc
	Recovery  gin.HandlerFunc
	Rate      gin.HandlerFunc
	// nanti tambah lagi misalnya Product, Order, dsb.
}

func RegisterRoutes(r *gin.Engine, h *Handlers, m *Middleware) {

	r.Use(m.CORS, m.RequestID, m.Logger, m.Recovery)

	mainGroup := r.Group("/")
	api := r.Group("/api/v1")

	protected := api.Group("")
	protected.Use(m.AuthJWT) // tinggal pakai

	// public route
	h.AuthSSO.RegisterRoute(mainGroup)

	// protected route
	h.User.RegisterRoute(protected)

	h.Mahasiswa.RegisterRoute(protected)

	api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"success": true, "status": "ok"}) })

}
