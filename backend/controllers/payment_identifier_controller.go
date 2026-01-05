package controllers

import (
	"net/http"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/services"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
)

// TriggerPaymentIdentifierWorker POST /api/v1/payment-identifier/trigger
// Trigger manual worker untuk check payment by identifier
func TriggerPaymentIdentifierWorker(c *gin.Context) {
	epnbpRepo := repositories.NewEpnbpRepository(database.DB)
	tagihanRepo := repositories.NewTagihanRepository(database.DB, database.DBPNBP)
	
	worker := services.NewPaymentIdentifierWorker(
		database.DB,
		database.DBPNBP,
		epnbpRepo,
		*tagihanRepo,
	)

	// Run worker secara synchronous untuk testing
	go func() {
		if err := worker.CheckAndUpdatePaymentByIdentifier(); err != nil {
			utils.Log.Errorf("Error in manual trigger: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Worker triggered successfully",
		"status":  "processing",
	})
}


