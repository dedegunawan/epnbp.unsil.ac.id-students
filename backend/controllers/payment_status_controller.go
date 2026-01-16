package controllers

import (
	"time"

	"net/http"
	"strconv"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getInvoiceInfoFromMySQL mengambil informasi invoice dari MySQL (DBPNBP)
// Returns: status, virtual_account, virtual_account_created_at
func getInvoiceInfoFromMySQL(invoiceID uint) (string, string, string) {
	type InvoiceResult struct {
		Status                string     `gorm:"column:status"`
		VirtualAccount        string     `gorm:"column:virtual_account"`
		VirtualAccountCreated *time.Time `gorm:"column:va_created_at"`
	}

	var result InvoiceResult

	// Query untuk ambil status invoice dan virtual account
	err := database.DBPNBP.
		Table("invoices").
		Select("invoices.status, virtual_accounts.virtual_account, virtual_accounts.created_at as va_created_at").
		Joins("LEFT JOIN virtual_accounts ON virtual_accounts.invoice_id = invoices.id").
		Where("invoices.id = ?", invoiceID).
		Order("virtual_accounts.created_at DESC").
		First(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Coba query hanya invoice tanpa join (jika virtual_account tidak ada)
			var invoiceOnly struct {
				Status string `gorm:"column:status"`
			}
			err2 := database.DBPNBP.
				Table("invoices").
				Select("status").
				Where("id = ?", invoiceID).
				First(&invoiceOnly).Error

			if err2 == nil {
				return invoiceOnly.Status, "", ""
			}
			return "", "", ""
		}
		return "", "", ""
	}

	var vaCreatedAt string
	if result.VirtualAccountCreated != nil {
		vaCreatedAt = result.VirtualAccountCreated.Format("2006-01-02 15:04:05")
	}

	return result.Status, result.VirtualAccount, vaCreatedAt
}

// PaymentStatusResponse response untuk status pembayaran
type PaymentStatusResponse struct {
	TotalBills   int64                 `json:"total_bills"`
	PaidBills    int64                 `json:"paid_bills"`
	UnpaidBills  int64                 `json:"unpaid_bills"`
	TotalAmount  int64                 `json:"total_amount"`
	PaidAmount   int64                 `json:"paid_amount"`
	UnpaidAmount int64                 `json:"unpaid_amount"`
	PaidList     []PaymentStatusDetail `json:"paid_list"`
	UnpaidList   []PaymentStatusDetail `json:"unpaid_list"`
}

// PaymentStatusDetail detail status pembayaran per tagihan
type PaymentStatusDetail struct {
	StudentBillID           uint   `json:"student_bill_id"`
	StudentID               string `json:"student_id"`
	StudentName             string `json:"student_name,omitempty"`
	AcademicYear            string `json:"academic_year"`
	BillName                string `json:"bill_name"`
	Amount                  int64  `json:"amount"`
	PaidAmount              int64  `json:"paid_amount"`
	RemainingAmount         int64  `json:"remaining_amount"`
	Status                  string `json:"status"`                               // "paid", "unpaid", "partial"
	StatusPostgreSQL        string `json:"status_postgresql"`                    // Status di PostgreSQL
	StatusDBPNBP            string `json:"status_dbpnbp,omitempty"`              // Status di MySQL DBPNBP
	VirtualAccount          string `json:"virtual_account,omitempty"`            // Virtual account number
	PayUrlCreatedAt         string `json:"pay_url_created_at,omitempty"`         // Tanggal dibuat pay_url
	VirtualAccountCreatedAt string `json:"virtual_account_created_at,omitempty"` // Tanggal dibuat virtual account
	PayUrlExpiredAt         string `json:"pay_url_expired_at,omitempty"`         // Tanggal expired pay_url
	InvoiceID               uint   `json:"invoice_id,omitempty"`                 // Invoice ID
	CreatedAt               string `json:"created_at"`
	UpdatedAt               string `json:"updated_at"`
}

// GetPaymentStatus GET /api/v1/payment-status
// Menampilkan status pembayaran: mana yang sudah bayar dan mana yang belum
func GetPaymentStatus(c *gin.Context) {
	// Query parameters
	studentID := c.Query("student_id")
	academicYear := c.Query("academic_year")
	status := c.Query("status") // "paid", "unpaid", "all"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	// Query builder
	query := database.DBPNBP.Model(&models.StudentBill{})

	// Filter by student_id
	if studentID != "" {
		query = query.Where("student_id = ?", studentID)
	}

	// Filter by academic_year
	if academicYear != "" {
		query = query.Where("academic_year = ?", academicYear)
	}

	// Get total count
	var totalCount int64
	query.Count(&totalCount)

	// Get bills with pagination
	var bills []models.StudentBill
	query = query.Preload("Mahasiswa").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	// Filter by status
	if status == "paid" {
		query = query.Where("(quantity * amount) - paid_amount <= 0")
	} else if status == "unpaid" {
		query = query.Where("(quantity * amount) - paid_amount > 0")
	}
	// else "all" - no additional filter

	if err := query.Find(&bills).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pembayaran"})
		return
	}

	// Get all student_bill_ids untuk query pay_urls
	var studentBillIDs []uint
	for _, bill := range bills {
		studentBillIDs = append(studentBillIDs, bill.ID)
	}

	// Get pay_urls untuk semua bills
	var payUrls []models.PayUrl
	if len(studentBillIDs) > 0 {
		database.DBPNBP.Where("student_bill_id IN ?", studentBillIDs).
			Order("created_at DESC").
			Find(&payUrls)
	}

	// Create map untuk quick lookup
	payUrlMap := make(map[uint]*models.PayUrl)
	for i := range payUrls {
		payUrlMap[payUrls[i].StudentBillID] = &payUrls[i]
	}

	// Process bills into paid and unpaid lists
	var paidList []PaymentStatusDetail
	var unpaidList []PaymentStatusDetail
	var totalAmount int64
	var paidAmount int64
	var unpaidAmount int64

	for _, bill := range bills {
		netAmount := bill.NetAmount()
		remaining := bill.Remaining()
		paid := bill.PaidAmount

		totalAmount += netAmount
		paidAmount += paid
		unpaidAmount += remaining

		detail := PaymentStatusDetail{
			StudentBillID:   bill.ID,
			StudentID:       bill.StudentID,
			AcademicYear:    bill.AcademicYear,
			BillName:        bill.Name,
			Amount:          netAmount,
			PaidAmount:      paid,
			RemainingAmount: remaining,
			CreatedAt:       bill.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:       bill.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// Get student name if available
		if bill.Mahasiswa != nil {
			detail.StudentName = bill.Mahasiswa.Nama
		}

		// Determine status PostgreSQL
		if remaining <= 0 {
			detail.Status = "paid"
			detail.StatusPostgreSQL = "paid"
			paidList = append(paidList, detail)
		} else if paid > 0 {
			detail.Status = "partial"
			detail.StatusPostgreSQL = "partial"
			unpaidList = append(unpaidList, detail)
		} else {
			detail.Status = "unpaid"
			detail.StatusPostgreSQL = "unpaid"
			unpaidList = append(unpaidList, detail)
		}

		// Get pay_url information
		if payUrl, exists := payUrlMap[bill.ID]; exists {
			detail.InvoiceID = payUrl.InvoiceID
			detail.PayUrlCreatedAt = payUrl.CreatedAt.Format("2006-01-02 15:04:05")
			if !payUrl.ExpiredAt.IsZero() {
				detail.PayUrlExpiredAt = payUrl.ExpiredAt.Format("2006-01-02 15:04:05")
			}

			// Get invoice status from MySQL (DBPNBP)
			detail.StatusDBPNBP, detail.VirtualAccount, detail.VirtualAccountCreatedAt = getInvoiceInfoFromMySQL(payUrl.InvoiceID)
		}
	}

	response := PaymentStatusResponse{
		TotalBills:   totalCount,
		PaidBills:    int64(len(paidList)),
		UnpaidBills:  int64(len(unpaidList)),
		TotalAmount:  totalAmount,
		PaidAmount:   paidAmount,
		UnpaidAmount: unpaidAmount,
		PaidList:     paidList,
		UnpaidList:   unpaidList,
	}

	c.JSON(http.StatusOK, response)
}

// GetPaymentStatusSummary GET /api/v1/payment-status/summary
// Menampilkan ringkasan status pembayaran
func GetPaymentStatusSummary(c *gin.Context) {
	academicYear := c.Query("academic_year")

	query := database.DBPNBP.Model(&models.StudentBill{})

	if academicYear != "" {
		query = query.Where("academic_year = ?", academicYear)
	}

	// Get summary statistics
	type SummaryResult struct {
		TotalBills   int64 `json:"total_bills"`
		PaidBills    int64 `json:"paid_bills"`
		UnpaidBills  int64 `json:"unpaid_bills"`
		PartialBills int64 `json:"partial_bills"`
		TotalAmount  int64 `json:"total_amount"`
		PaidAmount   int64 `json:"paid_amount"`
		UnpaidAmount int64 `json:"unpaid_amount"`
	}

	var summary SummaryResult

	// Count total bills
	query.Count(&summary.TotalBills)

	// Count paid bills (fully paid)
	query.Where("(quantity * amount) - paid_amount <= 0").Count(&summary.PaidBills)

	// Count unpaid bills (not paid at all)
	query.Where("paid_amount = 0 AND (quantity * amount) > 0").Count(&summary.UnpaidBills)

	// Count partial bills (partially paid)
	query.Where("paid_amount > 0 AND (quantity * amount) - paid_amount > 0").Count(&summary.PartialBills)

	// Calculate amounts
	var bills []models.StudentBill
	query.Find(&bills)

	for _, bill := range bills {
		netAmount := bill.NetAmount()
		summary.TotalAmount += netAmount
		summary.PaidAmount += bill.PaidAmount
		summary.UnpaidAmount += (netAmount - bill.PaidAmount)
	}

	c.JSON(http.StatusOK, summary)
}

// UpdatePaymentStatusRequest request untuk update status pembayaran
type UpdatePaymentStatusRequest struct {
	PaidAmount   int64  `json:"paid_amount" binding:"required"`   // Jumlah yang dibayar (bisa partial atau full)
	PaymentDate  string `json:"payment_date"`                     // Format: "2006-01-02 15:04:05" atau "2006-01-02"
	PaymentMethod string `json:"payment_method"`                   // VA, Transfer, Tunai, etc.
	Bank         string `json:"bank"`                              // Nama bank (optional)
	PaymentRef   string `json:"payment_ref"`                       // Referensi pembayaran (optional)
	Note         string `json:"note"`                              // Catatan (optional)
}

// UpdatePaymentStatus PUT /api/v1/payment-status/:id
// Mengupdate status pembayaran untuk tagihan tertentu
func UpdatePaymentStatus(c *gin.Context) {
	studentBillIDStr := c.Param("id")
	studentBillID, err := strconv.ParseUint(studentBillIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student_bill_id"})
		return
	}

	var req UpdatePaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ambil student bill
	var studentBill models.StudentBill
	if err := database.DBPNBP.Preload("Mahasiswa").First(&studentBill, uint(studentBillID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Student bill not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data tagihan"})
		return
	}

	// Validasi paid_amount
	netAmount := studentBill.NetAmount()
	if req.PaidAmount < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paid_amount tidak boleh negatif"})
		return
	}

	if req.PaidAmount > netAmount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paid_amount tidak boleh lebih besar dari net_amount"})
		return
	}

	// Tidak menyimpan ke database - hanya consume data dari DBPNBP (read-only)
	// Operasi write dihapus - paymentDate, paymentMethod, paymentRef, oldPaidAmount tidak digunakan

	// Reload student bill dengan data terbaru
	database.DBPNBP.Preload("Mahasiswa").First(&studentBill, uint(studentBillID))

	// Hitung status baru
	netAmount = studentBill.NetAmount()
	remaining := studentBill.Remaining()
	status := "unpaid"
	if remaining <= 0 {
		status = "paid"
	} else if studentBill.PaidAmount > 0 {
		status = "partial"
	}

	// Response
	response := gin.H{
		"message": "Status pembayaran berhasil diupdate",
		"data": gin.H{
			"student_bill_id":   studentBill.ID,
			"student_id":        studentBill.StudentID,
			"student_name":      "",
			"academic_year":     studentBill.AcademicYear,
			"bill_name":         studentBill.Name,
			"amount":            netAmount,
			"paid_amount":       studentBill.PaidAmount,
			"remaining_amount":  remaining,
			"status":            status,
			"updated_at":        studentBill.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	if studentBill.Mahasiswa != nil {
		response["data"].(gin.H)["student_name"] = studentBill.Mahasiswa.Nama
	}

	c.JSON(http.StatusOK, response)
}
