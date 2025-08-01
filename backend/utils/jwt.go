package utils

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"os"
	"time"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Name   string    `json:"name"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateJWT(userID uuid.UUID, email string, name string, t time.Time) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(t),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func CheckJwt(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Pastikan token menggunakan algoritma HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Validasi waktu expire
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}
func CheckJwtBySecret(tokenStr string, secret []byte, timeValidation bool) (map[string]interface{}, error) {
	// Gunakan jwt.MapClaims agar bisa fleksibel
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Pastikan algoritma yang digunakan adalah HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	// Ambil dan validasi claims sebagai MapClaims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Validasi waktu expire jika diperlukan
	if timeValidation {
		if expRaw, ok := claims["exp"]; ok {
			switch exp := expRaw.(type) {
			case float64:
				if int64(exp) < time.Now().Unix() {
					return nil, errors.New("token expired")
				}
			case json.Number:
				// jika token dari sumber lain kadang exp disimpan sebagai json.Number
				expInt, _ := exp.Int64()
				if expInt < time.Now().Unix() {
					return nil, errors.New("token expired")
				}
			}
		}
	}

	// Kembalikan claims sebagai map[string]interface{}
	return claims, nil
}
