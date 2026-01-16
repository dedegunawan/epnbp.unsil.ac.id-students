package controllers

import (
	"net/http"
	"strconv"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
)

// StudentBillDetail detail tagihan mahasiswa
type StudentBillDetail struct {
	ID                uint   `json:"id"`
	StudentID         string `json:"student_id"`
	StudentName       string `json:"student_name,omitempty"`
	AcademicYear      string `json:"academic_year"`
	BillName          string `json:"bill_name"`
	Quantity          int    `json:"quantity"`
	Amount            int64  `json:"amount"`
	Beasiswa          int64  `json:"beasiswa"`
	PaidAmount        int64  `json:"paid_amount"`
	RemainingAmount   int64  `json:"remaining_amount"`
	NetAmount         int64  `json:"net_amount"`
	Status            string `json:"status"` // "paid", "unpaid", "partial"
	Draft             bool   `json:"draft"`
	Note              string `json:"note"`
	InvoiceID         *uint  `json:"invoice_id,omitempty"`         // Invoice ID dari PNBP
	VirtualAccount    string `json:"virtual_account,omitempty"`     // Virtual Account dari PNBP
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// StudentBillsResponse response untuk semua tagihan mahasiswa
type StudentBillsResponse struct {
	TotalBills   int64               `json:"total_bills"`
	PaidBills    int64               `json:"paid_bills"`
	UnpaidBills  int64               `json:"unpaid_bills"`
	PartialBills int64               `json:"partial_bills"`
	TotalAmount  int64               `json:"total_amount"`
	PaidAmount   int64               `json:"paid_amount"`
	UnpaidAmount int64               `json:"unpaid_amount"`
	Bills        []StudentBillDetail `json:"bills"`
	Pagination   PaginationInfo      `json:"pagination"`
}

// PaginationInfo informasi pagination
type PaginationInfo struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	TotalPages  int   `json:"total_pages"`
	TotalItems  int64 `json:"total_items"`
	HasNext     bool  `json:"has_next"`
	HasPrev     bool  `json:"has_prev"`
}

// GetAllStudentBills GET /api/v1/student-bills
// Menampilkan semua tagihan mahasiswa dengan status pembayaran
func GetAllStudentBills(c *gin.Context) {
	// Query parameters
	studentID := c.Query("student_id")
	academicYear := c.Query("academic_year")
	status := c.Query("status") // "paid", "unpaid", "partial", "all"
	search := c.Query("search") // Search by student name, bill name, student_id
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

	// Query builder - hanya ambil dari finance year yang aktif
	// Join dengan mahasiswa untuk search by name
	baseQuery := database.DBPNBP.Model(&models.StudentBill{}).
		Joins("INNER JOIN finance_years ON finance_years.academic_year = student_bills.academic_year").
		Joins("LEFT JOIN mahasiswas ON mahasiswas.mhsw_id = student_bills.student_id").
		Where("finance_years.is_active = ?", true)

	// Filter by student_id
	if studentID != "" {
		baseQuery = baseQuery.Where("student_bills.student_id = ?", studentID)
	}

	// Filter by academic_year
	if academicYear != "" {
		baseQuery = baseQuery.Where("student_bills.academic_year = ?", academicYear)
	}

	// Search functionality
	if search != "" {
		searchPattern := "%" + search + "%"
		baseQuery = baseQuery.Where(
			"student_bills.student_id ILIKE ? OR "+
				"student_bills.name ILIKE ? OR "+
				"COALESCE(mahasiswas.nama, '') ILIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// Get total count (before status filter for accurate count)
	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		utils.Log.Error("GetAllStudentBills: Failed to count bills", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghitung total tagihan",
			"message": err.Error(),
		})
		return
	}

	// Apply status filter for query
	query := baseQuery

	// Get bills with pagination
	var bills []models.StudentBill
	query = query.Preload("Discounts").
		Order("student_bills.created_at DESC").
		Limit(limit).
		Offset(offset)

	// Note: Status filtering akan dilakukan setelah data diambil karena perlu menghitung NetAmount() dan Remaining()
	// Filter sederhana berdasarkan paid_amount untuk optimasi
	if status == "paid" {
		// Approximate: jika paid_amount >= amount (tanpa discount), kemungkinan sudah lunas
		query = query.Where("student_bills.paid_amount >= student_bills.quantity * student_bills.amount")
	} else if status == "unpaid" {
		query = query.Where("student_bills.paid_amount = 0 AND (student_bills.quantity * student_bills.amount) > 0")
	} else if status == "partial" {
		query = query.Where("student_bills.paid_amount > 0 AND student_bills.paid_amount < student_bills.quantity * student_bills.amount")
	}
	// else "all" - no additional filter

	if err := query.Find(&bills).Error; err != nil {
		utils.Log.Error("GetAllStudentBills: Failed to fetch bills", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data tagihan",
			"message": err.Error(),
		})
		return
	}

	// Maps untuk menyimpan data PNBP
	payUrlMap := make(map[uint]*models.PayUrl)
	vaMap := make(map[uint]string)

	// Load Mahasiswa separately untuk menghindari masalah relasi
	if len(bills) > 0 {
		var studentIDs []string
		var billIDs []uint
		for _, bill := range bills {
			if bill.StudentID != "" {
				studentIDs = append(studentIDs, bill.StudentID)
			}
			billIDs = append(billIDs, bill.ID)
		}
		
		// Load Mahasiswa
		if len(studentIDs) > 0 {
			var mahasiswas []models.Mahasiswa
			if err := database.DBPNBP.Where("mhsw_id IN ?", studentIDs).Find(&mahasiswas).Error; err == nil {
				// Map mahasiswa by mhsw_id
				mahasiswaMap := make(map[string]*models.Mahasiswa)
				for i := range mahasiswas {
					mahasiswaMap[mahasiswas[i].MhswID] = &mahasiswas[i]
				}
				
				// Assign mahasiswa to bills
				for i := range bills {
					if mahasiswa, ok := mahasiswaMap[bills[i].StudentID]; ok {
						bills[i].Mahasiswa = mahasiswa
					}
				}
			}
		}

		// Load PayUrl untuk mendapatkan invoice_id
		if len(billIDs) > 0 {
			var payUrls []models.PayUrl
			if err := database.DBPNBP.Where("student_bill_id IN ?", billIDs).
				Order("created_at DESC").Find(&payUrls).Error; err == nil {
				// Map payUrl by student_bill_id (ambil yang terbaru jika ada multiple)
				for i := range payUrls {
					if _, exists := payUrlMap[payUrls[i].StudentBillID]; !exists {
						payUrlMap[payUrls[i].StudentBillID] = &payUrls[i]
					}
				}
			}
		}

		// Load PaymentConfirmation untuk mendapatkan virtual_account
		if len(billIDs) > 0 {
			var paymentConfirmations []models.PaymentConfirmation
			if err := database.DBPNBP.Where("student_bill_id IN ?", billIDs).
				Order("created_at DESC").Find(&paymentConfirmations).Error; err == nil {
				// Map va_number by student_bill_id (ambil yang terbaru jika ada multiple)
				for _, pc := range paymentConfirmations {
					if _, exists := vaMap[pc.StudentBillID]; !exists && pc.VaNumber != "" {
						vaMap[pc.StudentBillID] = pc.VaNumber
					}
				}
			}
		}
	}

	// Calculate summary from ALL data (not just paginated)
	// Query semua bills untuk summary (tanpa pagination, TANPA filter status)
	// Summary harus menampilkan total dari semua status (paid, unpaid, partial)
	var allBillsForSummary []models.StudentBill
	summaryQuery := baseQuery.Preload("Discounts")
	// JANGAN apply status filter untuk summary - summary harus dari semua data
	
	if err := summaryQuery.Find(&allBillsForSummary).Error; err != nil {
		utils.Log.Error("GetAllStudentBills: Failed to fetch bills for summary", "error", err.Error())
		// Continue dengan data yang sudah ada, tapi summary akan 0
		allBillsForSummary = []models.StudentBill{}
	}

	// Calculate summary from all bills
	var paidBills int64
	var unpaidBills int64
	var partialBills int64
	var totalAmount int64
	var paidAmount int64
	var unpaidAmount int64

	for _, bill := range allBillsForSummary {
		netAmount := bill.NetAmount()
		remaining := bill.Remaining()
		paid := bill.PaidAmount

		totalAmount += netAmount
		paidAmount += paid
		unpaidAmount += remaining

		// Determine status
		if remaining <= 0 {
			paidBills++
		} else if paid > 0 {
			partialBills++
		} else {
			unpaidBills++
		}
	}

	// Process bills untuk ditampilkan (dengan pagination)
	var billDetails []StudentBillDetail

	for _, bill := range bills {
		netAmount := bill.NetAmount()
		remaining := bill.Remaining()
		paid := bill.PaidAmount

		// Determine status
		var billStatus string
		if remaining <= 0 {
			billStatus = "paid"
		} else if paid > 0 {
			billStatus = "partial"
		} else {
			billStatus = "unpaid"
		}

		detail := StudentBillDetail{
			ID:              bill.ID,
			StudentID:       bill.StudentID,
			AcademicYear:    bill.AcademicYear,
			BillName:        bill.Name,
			Quantity:        bill.Quantity,
			Amount:          bill.Amount,
			Beasiswa:        bill.Beasiswa,
			PaidAmount:      paid,
			RemainingAmount: remaining,
			NetAmount:       netAmount,
			Status:          billStatus,
			Draft:           bill.Draft,
			Note:            bill.Note,
			CreatedAt:       bill.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:       bill.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// Get student name if available
		if bill.Mahasiswa != nil {
			detail.StudentName = bill.Mahasiswa.Nama
		}

		// Get invoice_id from PayUrl (PNBP)
		if payUrl, ok := payUrlMap[bill.ID]; ok && payUrl.InvoiceID > 0 {
			detail.InvoiceID = &payUrl.InvoiceID
		}

		// Get virtual_account from PaymentConfirmation (PNBP)
		if vaNumber, ok := vaMap[bill.ID]; ok {
			detail.VirtualAccount = vaNumber
		}

		billDetails = append(billDetails, detail)
	}

	// Pastikan Bills selalu array, tidak pernah nil
	if billDetails == nil {
		billDetails = []StudentBillDetail{}
	}

	// Calculate pagination info
	totalPages := int((totalCount + int64(limit) - 1) / int64(limit)) // Ceiling division
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

	response := StudentBillsResponse{
		TotalBills:   totalCount,
		PaidBills:    paidBills,
		UnpaidBills:  unpaidBills,
		PartialBills: partialBills,
		TotalAmount:  totalAmount,
		PaidAmount:   paidAmount,
		UnpaidAmount: unpaidAmount,
		Bills:        billDetails,
		Pagination:   pagination,
	}

	c.JSON(http.StatusOK, response)
}

