package controllers

import (
	"net/http"
	"strconv"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
)

// PaymentStatusLogDetail detail log perubahan status
type PaymentStatusLogDetail struct {
	ID                uint   `json:"id"`
	StudentBillID     uint   `json:"student_bill_id"`
	StudentID         string `json:"student_id"`
	OldStatus         string `json:"old_status"`
	NewStatus       string `json:"new_status"`
	OldPaidAmount    int64  `json:"old_paid_amount"`
	NewPaidAmount     int64  `json:"new_paid_amount"`
	Amount            int64  `json:"amount"`
	PaymentDate       string `json:"payment_date,omitempty"`
	InvoiceID         *uint  `json:"invoice_id,omitempty"`
	VirtualAccount    string `json:"virtual_account,omitempty"`
	Identifier        string `json:"identifier"`
	TimeDifference    int64  `json:"time_difference"` // dalam detik
	Source            string `json:"source"`
	Message           string `json:"message"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// PaymentStatusLogsResponse response untuk log perubahan status
type PaymentStatusLogsResponse struct {
	TotalLogs int64                    `json:"total_logs"`
	Logs      []PaymentStatusLogDetail `json:"logs"`
	Pagination PaginationInfo          `json:"pagination"`
}

// GetAllPaymentStatusLogs GET /api/v1/payment-status-logs
// Menampilkan semua log perubahan status pembayaran
func GetAllPaymentStatusLogs(c *gin.Context) {
	// Query parameters
	studentID := c.Query("student_id")
	studentBillID := c.Query("student_bill_id")
	identifier := c.Query("identifier")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Query builder
	query := database.DBPNBP.Model(&models.PaymentStatusLog{})

	// Filter by student_id
	if studentID != "" {
		query = query.Where("student_id = ?", studentID)
	}

	// Filter by student_bill_id
	if studentBillID != "" {
		if billID, err := strconv.ParseUint(studentBillID, 10, 32); err == nil {
			query = query.Where("student_bill_id = ?", uint(billID))
		}
	}

	// Filter by identifier
	if identifier != "" {
		query = query.Where("identifier = ?", identifier)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		utils.Log.Error("GetAllPaymentStatusLogs: Failed to count logs", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghitung total log",
			"message": err.Error(),
		})
		return
	}

	// Get logs with pagination
	var logs []models.PaymentStatusLog
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		utils.Log.Error("GetAllPaymentStatusLogs: Failed to fetch logs", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data log",
			"message": err.Error(),
		})
		return
	}

	// Process logs
	var logDetails []PaymentStatusLogDetail
	for _, log := range logs {
		detail := PaymentStatusLogDetail{
			ID:             log.ID,
			StudentBillID:  log.StudentBillID,
			StudentID:      log.StudentID,
			OldStatus:      log.OldStatus,
			NewStatus:      log.NewStatus,
			OldPaidAmount:  log.OldPaidAmount,
			NewPaidAmount:  log.NewPaidAmount,
			Amount:         log.Amount,
			InvoiceID:      log.InvoiceID,
			VirtualAccount: log.VirtualAccount,
			Identifier:     log.Identifier,
			TimeDifference: log.TimeDifference,
			Source:         log.Source,
			Message:        log.Message,
			CreatedAt:      log.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      log.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if log.PaymentDate != nil {
			detail.PaymentDate = log.PaymentDate.Format("2006-01-02 15:04:05")
		}

		logDetails = append(logDetails, detail)
	}

	// Calculate pagination info
	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}

	pagination := PaginationInfo{
		CurrentPage: page,
		PerPage:     limit,
		TotalPages:  totalPages,
		TotalItems:  totalCount,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
	}

	response := PaymentStatusLogsResponse{
		TotalLogs:  totalCount,
		Logs:       logDetails,
		Pagination: pagination,
	}

	c.JSON(http.StatusOK, response)
}






