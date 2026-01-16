package config

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	utils.Log.Info("Loading environment variables from env, ", os.Getenv("APP_ENV"))
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			utils.Log.Fatal("Warning: .env file not found")
		}
	}

	utils.Log.Info("Environment variables loaded")
}

func GetEnv(key string) string {
	return os.Getenv(key)
}

func GetDefaultTokenExpired() time.Time {
	return time.Now().Add(2 * time.Hour)
}

func GetDefaultRefreshTokenExpired() time.Time {
	return time.Now().Add(24 * time.Hour)
}

// GetEmailSuffix mengembalikan email suffix yang diizinkan dari environment variable
// Default: @student.unsil.ac.id
func GetEmailSuffix() string {
	suffix := os.Getenv("EMAIL_SUFFIX")
	if suffix == "" {
		return "@student.unsil.ac.id"
	}
	return suffix
}

// ValidateEmailSuffix memvalidasi apakah email memiliki suffix yang diizinkan
func ValidateEmailSuffix(email string) bool {
	suffix := GetEmailSuffix()
	return len(email) >= len(suffix) && email[len(email)-len(suffix):] == suffix
}
