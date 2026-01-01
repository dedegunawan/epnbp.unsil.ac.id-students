package middleware

import (
	"context"
	"errors"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/usecase"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/authoidc"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/jwtmanager"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ContextKey untuk set/get dari Gin Context
const (
	ContextUserID = "user_id"
	ContextEmail  = "user_email"
	ContextName   = "user_name"
	ContextSsoID  = "user_sso_id"
	ContextToken  = "user_token" // jika perlu menyimpan token di context
)

type JwtMiddleware struct {
	Mgr      *jwtmanager.Manager
	Logger   *logger.Logger
	AuthOidc *authoidc.AuthOidc
	usecases *usecase.Usecase
}

func NewJwtMiddleware(mgr *jwtmanager.Manager, logger *logger.Logger, usecases *usecase.Usecase, oidc *authoidc.AuthOidc) *JwtMiddleware {
	return &JwtMiddleware{
		Mgr:      mgr,
		Logger:   logger,
		usecases: usecases,
		AuthOidc: oidc,
	}
}

func (jt JwtMiddleware) AuthJWT() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 0. Cek apakah ada header Authorization
		jt.Logger.Info("AuthJWT")
		tokenStr := extractAccessToken(c)
		if tokenStr == "" {
			response.Error(c, http.StatusUnauthorized, "Missing or invalid Authorization header")
			c.Abort()
			return
		}

		// 1. cari di db
		jt.Logger.Info("Get Access Token from DB")
		userToken, err := jt.usecases.UserTokenUsecase.GetByAccessToken(tokenStr)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Token not recognized")
			c.Abort()
			return
		}

		// 2. Verifikasi token JWT
		jt.Logger.Info("Check JWT Token")
		var claims *jwtmanager.Claims
		if userToken.JwtType == entity.JWTTypeInternal {
			claims, err = jt.CheckInternalToken(tokenStr)
		} else {
			claims, err = jt.CheckKeycloackToken(tokenStr)
		}

		if err != nil {
			response.Error(c, http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		// 3. Cek apakah user masih aktif
		jt.Logger.Info("Check User Status")
		user, err := jt.usecases.UserUsecase.GetByID(userToken.UserID)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		isActive := user.IsActive
		if !isActive {
			response.Error(c, http.StatusUnauthorized, "User is not active")
			c.Abort()
			return
		}

		// 4. Tambahkan context user_id dan sso_id
		c.Set(ContextSsoID, user.SsoID)
		c.Set(ContextName, claims.Name)
		c.Set(ContextUserID, user.ID)
		c.Set(ContextEmail, claims.Email)
		c.Set(ContextToken, tokenStr)

		c.Next()
	}
}

func (jt JwtMiddleware) CheckKeycloackToken(tokenStr string) (*jwtmanager.Claims, error) {
	// 1. Verifikasi token OIDC menggunakan AuthOidc Verifier
	jt.Logger.Info("Check Keycloak Token")
	jt.Logger.Info("Context : ", jt.usecases.UserTokenUsecase.GetContext())
	jt.Logger.Info("Authoidc : ", jt.AuthOidc)
	jt.Logger.Info("Verifier : ", jt.AuthOidc.Verifier)
	idToken, err := jt.AuthOidc.Verifier.Verify(
		context.Background(),
		tokenStr,
	)
	jt.Logger.Info("Verify token email : ", idToken)
	if err != nil {
		return nil, errors.New("Invalid token")
	}

	// 2. Ambil SSO sub dari token (bisa juga name/email jika perlu)
	var claims jwtmanager.Claims
	if err := idToken.Claims(&claims); err != nil {
		jt.Logger.Info("Invalid token")
		return nil, errors.New("Invalid claims")
	}
	return &claims, nil
}

func (jt JwtMiddleware) CheckInternalToken(tokenStr string) (*jwtmanager.Claims, error) {
	claims, err := jt.Mgr.Validate(tokenStr)
	if err != nil {
		return nil, err
	}
	return claims, nil

}

func extractAccessToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Fallback ke cookie (untuk browser)
	cookie, err := c.Cookie("access_token")
	if err == nil {
		return cookie
	}

	return ""
}
