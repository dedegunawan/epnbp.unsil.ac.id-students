package main

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/config"
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/routes"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
)

func main() {
	utils.InitLogger()
	utils.Log.Info("🚀 Starting application...")

	config.LoadEnv()

	utils.InitStorage()

	database.ConnectDatabasePnbp()

	// Semua data hanya read-only dari DBPNBP (MySQL)
	// Tidak ada worker, tidak ada operasi write ke database

	r := routes.SetupRouter()

	appPort := config.GetEnv("APP_PORT")
	if appPort == "" {
		appPort = "8080" // Default port
		utils.Log.Warn("⚠️ APP_PORT not set, using default port 8080")
	}

	utils.Log.Infof("✅ Server running at :%s", appPort)
	err := r.Run(":" + appPort)
	if err != nil {
		utils.Log.Fatal("❌ Failed to start server:", err)
	}
}
