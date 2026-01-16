package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/coreos/go-oidc/v3/oidc"
	"github.com/dedegunawan/backend-ujian-telp-v5/auth"
	"github.com/dedegunawan/backend-ujian-telp-v5/config"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
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
	// Login dengan email/password tidak didukung - gunakan SSO login
	c.JSON(http.StatusNotFound, gin.H{"error": "Login dengan email/password tidak didukung, gunakan SSO login"})
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

	// Validasi email suffix dari SSO
	if claims.Email != "" && !config.ValidateEmailSuffix(claims.Email) {
		utils.ErrorHandler(c, http.StatusUnauthorized, fmt.Sprintf("Email harus menggunakan domain %s", config.GetEmailSuffix()))
		return
	}

	// Tidak perlu create/update user - hanya cek email suffix
	// Data mahasiswa langsung dari mahasiswa_masters (read-only)
	studentID := utils.GetEmailPrefix(claims.Email)
	utils.Log.Info(fmt.Sprintf("Login berhasil untuk studentID: %s, email: %s", studentID, claims.Email))

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
