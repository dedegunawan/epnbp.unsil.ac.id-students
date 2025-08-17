package auth

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/user"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type registerReq struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthHandler struct {
	svc user.Service
}

func NewAuthHandler(svc user.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	u, err := h.svc.Register(req.Name, req.Email, req.Password)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Created(c, u)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 1. Cari user di DB
	u, err := h.svc.GetByEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// 2. Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// 3. Buat token (sementara dummy, nanti bisa ganti JWT)
	token := "dummy-jwt-token-for-" + u.Email

	// 4. Kirim response
	response.OK(c, gin.H{"token": token})
}

func (h *AuthHandler) RegisterRoute(g *gin.RouterGroup) {
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
}
