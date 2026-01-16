package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dedegunawan/backend-ujian-telp-v5/auth"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
)

type Claims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func RequireAuthFromTokenDB() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractAccessToken(c)
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			c.Abort()
			return
		}

		// Verifikasi token langsung dari JWT - tidak perlu cek database (read-only)
		// Coba verifikasi sebagai Keycloak token dulu (SSO)
		claims, err := checkKeycloackToken(tokenStr)
		if err != nil {
			// Jika bukan Keycloak token, coba sebagai internal JWT
			claims, err = checkInternalToken(tokenStr)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
		}

		// Tambahkan context dari claims
		c.Set("sso_id", claims.Sub)
		c.Set("email", claims.Email)
		c.Set("name", claims.Name)

		// Log untuk debugging
		utils.Log.Info("Auth middleware - Token verified", map[string]interface{}{
			"sso_id": claims.Sub,
			"email":  claims.Email,
			"name":   claims.Name,
		})

		// Untuk internal JWT, user_id ada di Sub
		if claims.Sub != "" {
			c.Set("user_id", claims.Sub)
		}

		c.Next()
	}
}

func checkKeycloackToken(tokenStr string) (*Claims, error) {
	idToken, err := auth.Verifier.Verify(context.Background(), tokenStr)
	utils.Log.Info("token")
	if err != nil {
		return nil, errors.New("Invalid token")
	}

	// 2. Ambil SSO sub dari token (bisa juga name/email jika perlu)
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		utils.Log.Info("Invalid token")
		return nil, errors.New("Invalid claims")
	}
	return &claims, nil
}

func checkInternalToken(tokenStr string) (*Claims, error) {
	claims, err := utils.CheckJwt(tokenStr)
	if err != nil {
		return nil, err
	}
	return &Claims{
		Sub:   claims.UserID.String(),
		Email: claims.Email,
		Name:  claims.Name,
	}, nil

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
