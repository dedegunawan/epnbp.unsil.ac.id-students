package main

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/config"
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/routes"
	"github.com/dedegunawan/backend-ujian-telp-v5/services"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
)

func main() {
	utils.InitLogger()
	utils.Log.Info("🚀 Starting application...")

	config.LoadEnv()

	utils.InitStorage()

	database.ConnectDatabase()
	database.ConnectDatabasePnbp()

	// Initialize repositories
	epnbpRepo := repositories.NewEpnbpRepository(database.DB)
	tagihanRepo := repositories.NewTagihanRepository(database.DB, database.DBPNBP)
	masterTagihanRepo := repositories.MasterTagihanRepository{DB: database.DBPNBP}
	tagihanService := services.NewTagihanService(*tagihanRepo, masterTagihanRepo)

	// Initialize Payment Status Worker
	paymentWorker := services.NewPaymentStatusWorker(
		database.DB,
		database.DBPNBP,
		epnbpRepo,
		*tagihanRepo,
		tagihanService,
	)

	// Start Payment Status Worker
	go paymentWorker.StartWorker("PaymentStatusWorker-1")
	utils.Log.Info("✅ Payment Status Worker started")

	// Initialize Payment Identifier Worker
	paymentIdentifierWorker := services.NewPaymentIdentifierWorker(
		database.DB,
		database.DBPNBP,
		epnbpRepo,
		*tagihanRepo,
	)

	// Start Payment Identifier Worker
	go paymentIdentifierWorker.StartWorker("PaymentIdentifierWorker-1")
	utils.Log.Info("✅ Payment Identifier Worker started")

	// Uncomment if you need other workers
	//worker := services.NewWorkerService(database.DB)
	//sintesys := services.NewSintesys(os.Getenv("SINTESYS_APP_URL"), os.Getenv("SINTESYS_APP_TOKEN"))
	//for i := 1; i <= 1; i++ {
	//	go worker.StartWorker(fmt.Sprintf("Worker-%d", i))
	//}
	//sintesys.ScanNewCallback()
	//for i := 1; i <= 2; i++ {
	//	go sintesys.ScanNewCallback()
	//}

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
