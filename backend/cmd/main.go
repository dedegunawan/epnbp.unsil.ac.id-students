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

	database.ConnectDatabase()
	r := routes.SetupRouter()
	utils.Log.Infof("✅ Server running at :%s", config.GetEnv("APP_PORT"))
	err := r.Run(":" + config.GetEnv("APP_PORT"))
	if err != nil {
		utils.Log.Fatal("❌ Failed to start server:", err)
	}
}
