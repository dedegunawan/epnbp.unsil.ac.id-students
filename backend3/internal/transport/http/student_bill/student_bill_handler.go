package student_bill

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/internal/domain/usecase"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/pkg/logger"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/pkg/mahasiswa_manager"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/pkg/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

type StudentBillHandler struct {
	logger   *logger.Logger
	usecases *usecase.Usecase
}

func NewStudentBillHandler(lg *logger.Logger, uc *usecase.Usecase) *StudentBillHandler {
	return &StudentBillHandler{logger: lg, usecases: uc}
}

// GetStudentBillStatus handles GET /api/v1/student-bill
func (h *StudentBillHandler) GetStudentBillStatus(c *gin.Context) {
	mahasiswaManager, err := mahasiswa_manager.NewFromContext(c, h.usecases, h.logger)
	if err != nil || mahasiswaManager == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create mahasiswa manager: "+err.Error())
		return
	}

	// Load tagihan
	mahasiswaManager.LoadTagihan()

	allPaid := mahasiswaManager.IsAllPaid()
	isGenerated := mahasiswaManager.IsTagihanGenerated()
	tagihanHarusDibayar := mahasiswaManager.TagihanHarusDibayar()
	historyTagihan := mahasiswaManager.HistoryTagihan()

	response := StudentBillResponse{
		Tahun:               mahasiswaManager.BudgetPeriod,
		IsPaid:              allPaid,
		IsGenerated:         isGenerated,
		TagihanHarusDibayar: tagihanHarusDibayar,
		HistoryTagihan:      historyTagihan,
	}

	c.JSON(http.StatusOK, response)
}

// GenerateCurrentBill handles POST /api/v1/student-bill
func (h *StudentBillHandler) GenerateCurrentBill(c *gin.Context) {
	mahasiswaManager, err := mahasiswa_manager.NewFromContext(c, h.usecases, h.logger)
	if err != nil || mahasiswaManager == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create mahasiswa manager: "+err.Error())
		return
	}

	mahasiswa := mahasiswaManager.Mahasiswa
	budgetPeriod := mahasiswaManager.BudgetPeriod

	// Check if mahasiswa is active
	// TODO: Implement GetIsMahasiswaAktifFromFullData logic

	// Check if mahasiswa is pascasarjana
	kodeProdi := mahasiswa.Prodi.KodeProdi
	isPasca := len(kodeProdi) >= 1 && (kodeProdi[0] == '8' || kodeProdi[0] == '9')

	var errGen error
	if isPasca {
		errGen = h.usecases.TagihanUsecase.CreateNewTagihanPasca(mahasiswa, budgetPeriod)
	} else {
		errGen = h.usecases.TagihanUsecase.CreateNewTagihan(mahasiswa, budgetPeriod)
	}

	if errGen != nil {
		h.logger.Error("Failed to generate bill", "error", errGen)
		response.Error(c, http.StatusInternalServerError, "Gagal membuat tagihan")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// RegenerateCurrentBill handles POST /api/v1/regenerate-student-bill
func (h *StudentBillHandler) RegenerateCurrentBill(c *gin.Context) {
	mahasiswaManager, err := mahasiswa_manager.NewFromContext(c, h.usecases, h.logger)
	if err != nil || mahasiswaManager == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create mahasiswa manager: "+err.Error())
		return
	}

	mhswID := mahasiswaManager.Mahasiswa.MhswID
	budgetPeriod := mahasiswaManager.BudgetPeriod

	// Delete unpaid bills
	err = h.usecases.TagihanUsecase.DeleteUnpaidBills(mhswID, budgetPeriod.AcademicYear)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menghapus tagihan")
		return
	}

	// Generate new bill
	h.GenerateCurrentBill(c)
}

// GenerateUrlPembayaran handles GET /api/v1/generate/:StudentBillID
func (h *StudentBillHandler) GenerateUrlPembayaran(c *gin.Context) {
	studentBillID := c.Param("StudentBillID")
	if studentBillID == "" {
		response.Error(c, http.StatusBadRequest, "StudentBillID is required")
		return
	}

	// Check if payment URL already exists and not expired
	payUrl, err := h.usecases.EpnbpUsecase.FindNotExpiredByStudentBill(studentBillID)
	if err == nil && payUrl != nil && payUrl.PayUrl != "" {
		c.JSON(http.StatusOK, payUrl)
		return
	}

	// Get mahasiswa from context
	mahasiswaManager, err := mahasiswa_manager.NewFromContext(c, h.usecases, h.logger)
	if err != nil || mahasiswaManager == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create mahasiswa manager: "+err.Error())
		return
	}

	// Get student bill
	studentBills, err := h.usecases.TagihanUsecase.GetStudentBills(mahasiswaManager.Mahasiswa.MhswID, mahasiswaManager.BudgetPeriod.AcademicYear)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil tagihan")
		return
	}

	var studentBill *entity.StudentBill
	for _, sb := range studentBills {
		if sb.ID == parseUint(studentBillID) {
			studentBill = &sb
			break
		}
	}

	if studentBill == nil {
		response.Error(c, http.StatusNotFound, "Tagihan tidak ditemukan")
		return
	}

	// Generate payment URL
	payUrl, err = h.usecases.EpnbpUsecase.GenerateNewPayUrl(
		*mahasiswaManager.User,
		*mahasiswaManager.Mahasiswa,
		*studentBill,
		mahasiswaManager.BudgetPeriod,
	)

	if err != nil {
		h.logger.Error("Failed to generate payment URL", "error", err)
		response.Error(c, http.StatusInternalServerError, "Gagal membuat URL Pembayaran")
		return
	}

	c.JSON(http.StatusOK, payUrl)
}

// ConfirmPembayaran handles POST /api/v1/confirm-payment/:StudentBillID
func (h *StudentBillHandler) ConfirmPembayaran(c *gin.Context) {
	studentBillID := c.Param("StudentBillID")
	if studentBillID == "" {
		response.Error(c, http.StatusBadRequest, "StudentBillID is required")
		return
	}

	vaNumber := c.PostForm("vaNumber")
	paymentDate := c.PostForm("paymentDate")

	if vaNumber == "" || paymentDate == "" {
		response.Error(c, http.StatusBadRequest, "Nomor VA dan Tanggal Bayar wajib diisi")
		return
	}

	// Get mahasiswa from context
	mahasiswaManager, err := mahasiswa_manager.NewFromContext(c, h.usecases, h.logger)
	if err != nil || mahasiswaManager == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create mahasiswa manager: "+err.Error())
		return
	}

	// Get student bill
	studentBills, err := h.usecases.TagihanUsecase.GetStudentBills(mahasiswaManager.Mahasiswa.MhswID, mahasiswaManager.BudgetPeriod.AcademicYear)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal mengambil tagihan")
		return
	}

	var studentBill *entity.StudentBill
	for _, sb := range studentBills {
		if sb.ID == parseUint(studentBillID) {
			studentBill = &sb
			break
		}
	}

	if studentBill == nil {
		response.Error(c, http.StatusNotFound, "Tagihan tidak ditemukan")
		return
	}

	// Handle file upload
	fileURL, ok := h.handleUpload(c, "file")
	if !ok {
		return // Error already handled in handleUpload
	}

	// Save payment confirmation
	paymentConfirmation, err := h.usecases.TagihanUsecase.SavePaymentConfirmation(
		*studentBill,
		vaNumber,
		paymentDate,
		fileURL,
	)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan konfirmasi pembayaran")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Bukti bayar berhasil dikirim",
		"studentBillID":       studentBill.ID,
		"vaNumber":            vaNumber,
		"paymentDate":         paymentDate,
		"fileURL":             fileURL,
		"paymentConfirmation": paymentConfirmation,
	})
}

// BackToSintesys handles GET /api/v1/back-to-sintesys
func (h *StudentBillHandler) BackToSintesys(c *gin.Context) {
	mahasiswaManager, err := mahasiswa_manager.NewFromContext(c, h.usecases, h.logger)
	if err != nil || mahasiswaManager == nil {
		// If error, just redirect to default URL
		h.redirectSintesys(c)
		return
	}

	mahasiswa := mahasiswaManager.Mahasiswa
	budgetPeriod := mahasiswaManager.BudgetPeriod

	// If UKT is 0, redirect immediately
	if mahasiswa.UKT == "0" {
		// TODO: Call SintesysService.SendCallback if needed
		h.redirectSintesys(c)
		return
	}

	// Check if all bills are paid
	studentBills, err := h.usecases.TagihanUsecase.GetStudentBills(mahasiswa.MhswID, budgetPeriod.AcademicYear)
	if err != nil {
		h.redirectSintesys(c)
		return
	}

	isPaid := false
	for _, studentBill := range studentBills {
		if studentBill.PaidAmount >= studentBill.Amount {
			isPaid = true
			break
		}
	}

	if isPaid {
		// TODO: Call SintesysService.SendCallback if needed
		h.redirectSintesys(c)
		return
	}

	h.redirectSintesys(c)
}

func (h *StudentBillHandler) redirectSintesys(c *gin.Context) {
	// TODO: Get from environment variable
	backUrl := "http://sintesys.unsil.ac.id"
	c.JSON(http.StatusOK, gin.H{"url": backUrl})
}

func (h *StudentBillHandler) handleUpload(c *gin.Context, filename string) (string, bool) {
	// TODO: Implement file upload to MinIO
	// This should be moved to a storage service
	fileHeader, err := c.FormFile(filename)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "File bukti bayar wajib diunggah")
		return "", false
	}

	// TODO: Upload to MinIO and return object name
	objectName := "bukti-bayar/" + fileHeader.Filename
	return objectName, true
}

func parseUint(s string) uint {
	// Simple parsing, should use strconv in production
	var result uint
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + uint(char-'0')
		}
	}
	return result
}

type StudentBillResponse struct {
	Tahun               entity.BudgetPeriod `json:"tahun"`
	IsPaid              bool                `json:"isPaid"`
	IsGenerated         bool                `json:"isGenerated"`
	TagihanHarusDibayar []entity.StudentBill `json:"tagihanHarusDibayar"`
	HistoryTagihan      []entity.StudentBill `json:"historyTagihan"`
}


