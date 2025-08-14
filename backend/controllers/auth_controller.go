package controllers

import (
	"context"
	"fmt"
	_ "github.com/coreos/go-oidc/v3/oidc"
	"github.com/dedegunawan/backend-ujian-telp-v5/auth"
	"github.com/dedegunawan/backend-ujian-telp-v5/config"
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/services"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
)

func SsoLoginHandler(c *gin.Context) {
	backTo := c.Query("backTo") // contoh: "/dashboard"
	state := utils.EncodeBackState(backTo)

	authCodeURL := auth.OAuth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, authCodeURL)
}

func SsoLogoutHandler(c *gin.Context) {
	redirectURI := os.Getenv("OIDC_LOGOUT_REDIRECT") // e.g. https://your-app.com
	clientID := os.Getenv("OIDC_CLIENT_ID")

	logoutURL := auth.GetLogoutURL(redirectURI, clientID)

	// Hapus session lokal (jika pakai cookie/token)

	// Redirect ke SSO logout endpoint
	c.Redirect(http.StatusTemporaryRedirect, logoutURL)
}

func LoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if !utils.BindJSONOrAbort(c, &req) {
		return
	}

	// Cari user berdasarkan email
	userRepo := repositories.UserRepository{DB: database.DB}

	user, err := userRepo.FindByEmail(req.Email)
	if err != nil {
		utils.ErrorHandler(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if user.Password == nil || len(*user.Password) <= 0 {
		empty := ""
		user.Password = &empty
	}

	utils.Log.Info("Login using email:", user.Email)

	// Validasi password (dengan bcrypt)
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password)); err != nil {
		utils.Log.Info(fmt.Sprintf("Login using email: %s, password: %s, hash: %s, err: %s", user.Email, req.Password, *user.Password, err.Error()))
		utils.ErrorHandler(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if user.IsActive == false {
		utils.ErrorHandler(c, http.StatusUnauthorized, "User is not active")
		return
	}

	utils.Log.Info("Try generate JWT by email :", user.Email)

	// Buat token JWT
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Name, config.GetDefaultTokenExpired())
	if err != nil {
		utils.ErrorHandler(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	userTokenRepo := repositories.UserTokenRepository{DB: database.DB}
	userTokenService := services.UserTokenService{Repo: &userTokenRepo, Context: c}

	newToken, err := userTokenService.SaveLoginUserToken(
		user.ID,
		token,
	)
	if err != nil {
		utils.ErrorHandler(c, http.StatusInternalServerError, "Failed to save token")
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

func RefreshHandler(c *gin.Context) {

}

func CallbackHandler(c *gin.Context) {
	ctx := context.Background()
	code := c.Query("code")

	_, err := utils.DecodeBackState(c.Query("state"))

	if code == "" {
		utils.ErrorHandler(c, http.StatusBadRequest, "No code provided")
		return
	}

	token, err := auth.OAuth2Config.Exchange(ctx, code)
	if err != nil {
		log.Println("❌ Token exchange failed:", err)
		utils.ErrorHandler(c, http.StatusUnauthorized, "Token exchange failed")
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		utils.ErrorHandler(c, http.StatusInternalServerError, "Missing id_token in token response")
		return
	}

	idToken, err := auth.Verifier.Verify(ctx, rawIDToken)
	//utils.Log.Info("Verify token email : ", idToken)
	if err != nil {
		utils.ErrorHandler(c, http.StatusUnauthorized, "Invalid ID Token")
		return
	}

	// Ambil klaim dari token
	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		utils.ErrorHandler(c, http.StatusInternalServerError, "Failed to parse claims")
		return
	}

	// Inisialisasi repository & service
	userRepo := repositories.UserRepository{DB: database.DB}
	userTokenRepo := repositories.UserTokenRepository{DB: database.DB}

	userService := services.UserService{Repo: &userRepo}
	userTokenService := services.UserTokenService{Repo: &userTokenRepo}

	// Temukan atau buat user berdasarkan sso_id
	user, err := userService.GetOrCreateByEmail(claims.Sub, claims.Email, claims.Name)
	if err != nil {
		utils.ErrorHandler(c, http.StatusInternalServerError, "Failed to create or find user")
		return
	}

	// sinkronkan dengan data npm simak
	mahasiswaRepo := repositories.NewMahasiswaRepository(database.DB)
	mahasiswaService := services.NewMahasiswaService(mahasiswaRepo)
	err = mahasiswaService.CreateFromSimak(utils.GetEmailPrefix(claims.Email))

	if err != nil {
		err = mahasiswaService.CreateFromMasterMahasiswa(utils.GetEmailPrefix(claims.Email))
	}

	if err != nil {
		utils.Log.Info(fmt.Sprintf("Err : %s", err.Error()))
		utils.ErrorHandler(c, http.StatusInternalServerError, "Failed to create mahasiswa user")
	}

	// Simpan access_token dan refresh_token
	err = userTokenService.SaveUserToken(
		user.ID,
		token.AccessToken,
		token.RefreshToken,
		token.TokenType,
		token.Expiry,
	)
	if err != nil {
		utils.ErrorHandler(c, http.StatusInternalServerError, "Failed to save token")
		return
	}

	frontendUrl := os.Getenv("FRONTEND_URL")
	accessToken := token.AccessToken

	utils.Log.Info("Redirect URL:", frontendUrl+"?token="+accessToken)
	c.Redirect(http.StatusFound, frontendUrl+"?token="+accessToken)

	// Respon sukses
	//utils.ErrorHandler(c, http.StatusOK, gin.H{
	//	"message":       "Login success",
	//	"user":          user,
	//	"access_token":  token.AccessToken,
	//	"refresh_token": token.RefreshToken,
	//	"expires_at":    token.Expiry,
	//})
}
