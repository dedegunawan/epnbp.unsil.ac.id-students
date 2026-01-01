package auth

import (
	"fmt"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/usecase"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/authoidc"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/encoder"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/jwtmanager"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/request"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/response"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/strings"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
)

type AuthSsoHandler struct {
	usecases usecase.Usecase
	auth     authoidc.AuthOidc
	lg       *logger.Logger
	jwt      jwtmanager.Manager
}

func NewAuthSsoHandler(usecases usecase.Usecase, auth authoidc.AuthOidc, lg *logger.Logger, jwt jwtmanager.Manager) *AuthSsoHandler {
	return &AuthSsoHandler{
		usecases: usecases,
		auth:     auth,
		lg:       lg,
		jwt:      jwt,
	}
}

func (h *AuthSsoHandler) SsoLoginHandler(c *gin.Context) {
	backTo := c.Query("backTo") // contoh: "/dashboard"
	state := encoder.EncodeBackState(backTo)

	authCodeURL := h.auth.OAuth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, authCodeURL)
}

func (h *AuthSsoHandler) SsoLogoutHandler(c *gin.Context) {

	logoutURL := h.auth.GetLogoutURL()

	// Hapus session lokal (jika pakai cookie/token)

	// Redirect ke SSO logout endpoint
	c.Redirect(http.StatusTemporaryRedirect, logoutURL)
}

func (h *AuthSsoHandler) LoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if !request.BindJSONOrAbort(c, &req) {
		return
	}

	user, err := h.usecases.UserUsecase.GetByEmail(req.Email)
	if err != nil {
		response.ErrorHandler(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	h.lg.Info("Login using email:", user.Email)

	// Validasi password (dengan bcrypt)
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		h.lg.Info(fmt.Sprintf("Login using email: %s, password: %s, hash: %s, err: %s", user.Email, req.Password, user.PasswordHash, err.Error()))
		response.ErrorHandler(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if user.IsActive == false {
		response.ErrorHandler(c, http.StatusUnauthorized, "User is not active")
		return
	}

	h.lg.Info("Try generate JWT by email :", user.Email)

	// Buat token JWT
	token, err := h.jwt.Generate(user.ID, user.Email, user.Name)
	if err != nil {
		response.ErrorHandler(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	newToken, err := h.usecases.UserTokenUsecase.SaveLoginUserToken(
		user.ID,
		token,
	)
	if err != nil {
		response.ErrorHandler(c, http.StatusInternalServerError, "Failed to save token")
		return
	}

	// Respon sukses
	c.JSON(http.StatusOK, gin.H{
		"message":       "Login success",
		"user":          user,
		"access_token":  newToken.AccessToken,
		"refresh_token": newToken.RefreshToken,
		"expires_at":    newToken.ExpiresAt,
	})
}

func (h *AuthSsoHandler) RefreshHandler(c *gin.Context) {

}

func (h *AuthSsoHandler) CallbackHandler(c *gin.Context) {
	code := c.Query("code")
	//h.userTokenService.SetContext(context.Background())

	_, err := encoder.DecodeBackState(c.Query("state"))

	if code == "" {
		response.ErrorHandler(c, http.StatusBadRequest, "No code provided")
		return
	}

	token, err := h.auth.OAuth2Config.Exchange(h.usecases.UserTokenUsecase.GetContext(), code)
	if err != nil {
		log.Println("❌ Token exchange failed:", err)
		response.ErrorHandler(c, http.StatusUnauthorized, "Token exchange failed")
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		response.ErrorHandler(c, http.StatusInternalServerError, "Missing id_token in token response")
		return
	}

	idToken, err := h.auth.Verifier.Verify(h.usecases.UserTokenUsecase.GetContext(), rawIDToken)
	//utils.Log.Info("Verify token email : ", idToken)
	if err != nil {
		response.ErrorHandler(c, http.StatusUnauthorized, "Invalid ID Token")
		return
	}

	// Ambil klaim dari token
	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		response.ErrorHandler(c, http.StatusInternalServerError, "Failed to parse claims")
		return
	}

	// Inisialisasi repository & service

	// Temukan atau buat user berdasarkan sso_id
	user, err := h.usecases.UserUsecase.GetOrCreateByEmail(claims.Sub, claims.Email, claims.Name)
	if err != nil {
		response.ErrorHandler(c, http.StatusInternalServerError, "Failed to create or find user")
		return
	}

	// ambil data mahasiswa simak

	// sinkronkan dengan data npm simak
	_, err = h.usecases.MahasiswaUsecase.FindOrSyncByStudentID(strings.GetEmailPrefix(claims.Email))

	if err != nil {
		h.lg.Info(fmt.Sprintf("Err : %s", err.Error()))
		response.ErrorHandler(c, http.StatusInternalServerError, "Failed to create mahasiswa user")
	}

	// Simpan access_token dan refresh_token
	err = h.usecases.UserTokenUsecase.SaveUserToken(
		user.ID,
		entity.JWTTypeKeycloak,
		token.AccessToken,
		token.RefreshToken,
		token.TokenType,
		token.Expiry,
	)
	if err != nil {
		response.ErrorHandler(c, http.StatusInternalServerError, "Failed to save token")
		return
	}

	frontendUrl := os.Getenv("FRONTEND_URL")
	accessToken := token.AccessToken

	h.lg.Info("Redirect URL:", frontendUrl+"?token="+accessToken)
	c.Redirect(http.StatusFound, frontendUrl+"?token="+accessToken)
}

func (h *AuthSsoHandler) RegisterRoute(r *gin.RouterGroup) {
	r.GET("/sso-login", h.SsoLoginHandler)
	r.GET("/sso-logout", h.SsoLogoutHandler)
	r.POST("/login", h.LoginHandler)
	r.GET("/callback", h.CallbackHandler)
}
