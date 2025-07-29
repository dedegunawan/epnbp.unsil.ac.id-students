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
