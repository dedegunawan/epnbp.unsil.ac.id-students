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
	database.ConnectDatabasePnbp()

	// Uncomment if you need to initialize repositories or services

	//worker := services.NewWorkerService(database.DB)
	//sintesys := services.NewSintesys(os.Getenv("SINTESYS_APP_URL"), os.Getenv("SINTESYS_APP_TOKEN"))

	// Jalankan 5 worker paralel
	//for i := 1; i <= 1; i++ {
	//go worker.StartWorker(fmt.Sprintf("Worker-%d", i))
	//}
	//sintesys.ScanNewCallback()
	//for i := 1; i <= 2; i++ {
	//	go sintesys.ScanNewCallback()
	//}

	r := routes.SetupRouter()
	utils.Log.Infof("✅ Server running at :%s", config.GetEnv("APP_PORT"))
	err := r.Run(":" + config.GetEnv("APP_PORT"))
	if err != nil {
		utils.Log.Fatal("❌ Failed to start server:", err)
	}
}
