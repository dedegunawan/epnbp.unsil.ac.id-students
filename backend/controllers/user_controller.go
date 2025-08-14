package controllers

import (
	"fmt"
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/services"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func getMahasiswa(c *gin.Context) (*models.User, *models.Mahasiswa, bool) {
	userRepo := repositories.UserRepository{DB: database.DB}
	ssoID := c.GetString("sso_id")
	utils.Log.Info("mahasiswa", "email", ssoID)
	user, err := userRepo.FindBySSOID(ssoID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, nil, true
	}
	utils.Log.Info("mahasiswa", "email")

	mahasiswaRepo := repositories.NewMahasiswaRepository(database.DB)
	email := user.Email
	mahasiswa, _ := mahasiswaRepo.FindByEmailPattern(email)

	utils.Log.Info("mahasiswa", "email", email, mahasiswa)

	return user, mahasiswa, false
}

func Me(c *gin.Context) {
	user, mahasiswa, mustreturn := getMahasiswa(c)
	if mustreturn {
		return
	}
	ssoID := c.GetString("sso_id")
	c.JSON(200, gin.H{
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"sso_id":    ssoID,
		"is_active": user.IsActive,
		"mahasiswa": mahasiswa,
	})
}

type StudentBillResponse struct {
	Tahun               models.FinanceYear   `json:"tahun"`
	IsPaid              bool                 `json:"isPaid"`
	IsGenerated         bool                 `json:"isGenerated"`
	TagihanHarusDibayar []models.StudentBill `json:"tagihanHarusDibayar"`
	HistoryTagihan      []models.StudentBill `json:"historyTagihan"`
}

// GET /student-bill
func GetStudentBillStatus(c *gin.Context) {
	_, mahasiswa, mustreturn := getMahasiswa(c)
	if mustreturn {
		return
	}

	mhswID := mahasiswa.MhswID

	tagihanRepo := repositories.TagihanRepository{DB: database.DB}

	// Panggil repository untuk ambil FinanceYear aktif
	activeYear, err := tagihanRepo.GetActiveFinanceYear()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tahun aktif tidak ditemukan"})
		return
	}

	// ambil tagihan mahasiswa semester sekarang
	tagihan, err := tagihanRepo.GetStudentBills(mhswID, activeYear.AcademicYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil tagihan"})
		return
	}

	// ambil tagihan yang semester sebelumnya belum dibayar
	unpaidTagihan, err := tagihanRepo.GetAllUnpaidBillsExcept(mhswID, activeYear.AcademicYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil tagihan"})
		return
	}

	// ambil tagihan semester sebelumnya yang sudah dibayar
	paidTagihan, err := tagihanRepo.GetAllPaidBillsExcept(mhswID, activeYear.AcademicYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil tagihan"})
		return
	}

	// Pisahkan: tagihan harus dibayar (belum lunas), dan histori
	var tagihanHarusDibayar []models.StudentBill
	var historyTagihan []models.StudentBill
	allPaid := true

	for _, t := range tagihan {
		if t.Remaining() > 0 {
			tagihanHarusDibayar = append(tagihanHarusDibayar, t)
			allPaid = false
		} else {
			historyTagihan = append(historyTagihan, t)
		}
	}

	for _, t := range unpaidTagihan {
		tagihanHarusDibayar = append(tagihanHarusDibayar, t)
	}

	for _, t := range paidTagihan {
		historyTagihan = append(historyTagihan, t)
	}

	isGenerated := len(tagihan) > 0
	if !isGenerated {
		allPaid = false
	}

	response := StudentBillResponse{
		Tahun:               *activeYear,
		IsPaid:              allPaid,
		IsGenerated:         isGenerated,
		TagihanHarusDibayar: tagihanHarusDibayar,
		HistoryTagihan:      historyTagihan,
	}

	c.JSON(http.StatusOK, response)
}

// POST /student-bill
func RegenerateCurrentBill(c *gin.Context) {
	_, mahasiswa, mustreturn := getMahasiswa(c)
	if mustreturn {
		return
	}

	mhswID := mahasiswa.MhswID
	tagihanRepo := repositories.TagihanRepository{DB: database.DB}

	// Panggil repository untuk ambil FinanceYear aktif
	activeYear, err := tagihanRepo.GetActiveFinanceYear()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tahun aktif tidak ditemukan"})
		return
	}

	err = tagihanRepo.DeleteUnpaidBills(mhswID, activeYear.AcademicYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil tagihan"})
		return
	}
	GenerateCurrentBill(c)
}

func GetIsMahasiswaAktifFromFullData(mahasiswa models.Mahasiswa) bool {
	full_data := mahasiswa.ParseFullData()
	statusMhswID, ok := full_data["StatusMhswID"].(string)
	if !ok || statusMhswID == "" {
		statusMhswID = "-"
	}

	statusAkademikId, ok := full_data["StatusAkademikID"].(int)
	if !ok {
		statusAkademikId = 0
	}
	utils.Log.Info("GetIsMahasiswaAktifFromFullData", map[string]string{
		"MhswID":       statusMhswID,
		"AkademikID":   fmt.Sprintf("%d", statusAkademikId),
		"statusMhswID": statusMhswID,
		"full_data":    fmt.Sprintf("%v", full_data),
		"lastOk":       fmt.Sprintf("%t", ok),
	})
	return statusMhswID == "A" || statusAkademikId == 1
}

func GenerateCurrentBill(c *gin.Context) {
	_, mahasiswa, mustreturn := getMahasiswa(c)
	if mustreturn {
		return
	}

	mhswID := mahasiswa.MhswID

	if len(mhswID) >= 3 {
		if mhswID[2] == '8' || mhswID[2] == '9' {
			GenerateCurrentBillPascasarjana(c, *mahasiswa)
			return
		}
	}

	tagihanRepo := repositories.TagihanRepository{DB: database.DB}

	// Panggil repository untuk ambil FinanceYear aktif
	activeYear, err := tagihanRepo.GetActiveFinanceYear()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tahun aktif tidak ditemukan"})
		return
	}

	// Panggil repository untuk ambil FinanceYear aktif
	// hardcode status mahasiswa aktif
	if !GetIsMahasiswaAktifFromFullData(*mahasiswa) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Pembuatan tagihan baru untuk tahun aktif, hanya diperboleh untuk mahasiswa aktif"})
		return
	}

	// Panggil repository untuk ambil tagihan mahasiswa
	tagihan, err := tagihanRepo.GetStudentBills(mhswID, activeYear.AcademicYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil tagihan"})
		return
	}

	tagihanService := services.NewTagihanSerice(tagihanRepo)

	if len(tagihan) == 0 {
		if err := tagihanService.CreateNewTagihan(mahasiswa, activeYear); err != nil {
			utils.Log.Info("Gagal membuat tagihan", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat tagihan"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GenerateCurrentBillPascasarjana(c *gin.Context, mahasiswa models.Mahasiswa) {
	utils.Log.Info("GenerateCurrentBillPascasarjana")
	mhswID := mahasiswa.MhswID
	tagihanRepo := repositories.TagihanRepository{DB: database.DB}

	// Panggil repository untuk ambil FinanceYear aktif
	activeYear, err := tagihanRepo.GetActiveFinanceYear()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tahun aktif tidak ditemukan"})
		return
	}

	// Panggil repository untuk ambil FinanceYear aktif
	if !GetIsMahasiswaAktifFromFullData(mahasiswa) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Pembuatan tagihan baru untuk tahun aktif, hanya diperboleh untuk mahasiswa aktif"})
		return
	}

	// Panggil repository untuk ambil tagihan mahasiswa
	tagihan, err := tagihanRepo.GetStudentBills(mhswID, activeYear.AcademicYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil tagihan"})
		return
	}

	tagihanService := services.NewTagihanSerice(tagihanRepo)

	if len(tagihan) == 0 {
		if err := tagihanService.CreateNewTagihanPasca(&mahasiswa, activeYear); err != nil {
			utils.Log.Info("Gagal membuat tagihan", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat tagihan"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GenerateUrlPembayaran(c *gin.Context) {
	studentBillID := c.Param("StudentBillID")
	if studentBillID == "" {
		utils.Log.Info("StudentBillID not found ", studentBillID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat URL Pembayaran"})
		return
	}

	epnbpRepo := repositories.NewEpnbpRepository(database.DB)

	payUrl, _ := epnbpRepo.FindNotExpiredByStudentBill(studentBillID)
	if payUrl != nil && payUrl.PayUrl != "" {
		c.JSON(http.StatusOK, payUrl)
		return
	}

	user, mahasiswa, mustreturn := getMahasiswa(c)
	if mustreturn {
		return
	}

	tagihanRepo := repositories.TagihanRepository{DB: database.DB}

	// Panggil repository untuk ambil FinanceYear aktif
	activeYear, err := tagihanRepo.GetActiveFinanceYear()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tahun aktif tidak ditemukan"})
		return
	}

	studentBill, err := tagihanRepo.FindStudentBillByID(studentBillID)
	if err != nil {
		utils.Log.Info("Gagal membuat tagihan", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat URL Pembayaran"})
		return
	}

	epnbpService := services.NewEpnbpService(epnbpRepo)
	payUrl, err = epnbpService.GenerateNewPayUrl(
		*user,
		*mahasiswa,
		*studentBill,
		*activeYear,
	)

	if err != nil {
		utils.Log.Info("NewEpnbpService error ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat URL Pembayaran"})
		return
	}

	c.JSON(http.StatusOK, payUrl)
	return

}

func ConfirmPembayaran(c *gin.Context) {
	studentBillID := c.Param("StudentBillID")
	if studentBillID == "" {
		utils.Log.Info("StudentBillID not found ", studentBillID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "StudentBillID wajib diisi"})
		return
	}

	// Ambil form input
	vaNumber := c.PostForm("vaNumber")
	paymentDate := c.PostForm("paymentDate")

	if vaNumber == "" || paymentDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor VA dan Tanggal Bayar wajib diisi"})
		return
	}

	// Validasi student bill (opsional)
	tagihanRepo := repositories.TagihanRepository{DB: database.DB}
	studentBill, err := tagihanRepo.FindStudentBillByID(studentBillID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tagihan tidak ditemukan"})
		return
	}

	fileURL, ok := handleUpload(c, "file")
	if !ok {
		return
	}

	// Simpan ke database (opsional, sesuaikan dengan struktur Anda)
	paymentConfirmation, err := services.NewTagihanSerice(tagihanRepo).SavePaymentConfirmation(*studentBill, vaNumber, paymentDate, fileURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan konfirmasi pembayaran"})
		return
	}

	// Sukses
	c.JSON(http.StatusOK, gin.H{
		"message":             "Bukti bayar berhasil dikirim",
		"studentBillID":       studentBill.ID,
		"vaNumber":            vaNumber,
		"paymentDate":         paymentDate,
		"fileURL":             fileURL,
		"paymentConfirmation": paymentConfirmation,
	})
}

func handleUpload(c *gin.Context, filename string) (string, bool) {
	// Ambil file dari form
	fileHeader, err := c.FormFile(filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File bukti bayar wajib diunggah"})
		return "", false
	}

	// Buka file dan baca kontennya sebagai []byte
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca file"})
		return "", false
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca isi file"})
		return "", false
	}

	// Tentukan object name unik untuk penyimpanan di MinIO
	ext := filepath.Ext(fileHeader.Filename)
	objectName := fmt.Sprintf("bukti-bayar/%s-%d%s", time.Now().Format("20060102-150405"), rand.Intn(99999), ext)

	// Upload ke MinIO
	_, err = utils.UploadObjectToMinio(objectName, fileBytes, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		utils.Log.Error("Gagal upload ke MinIO", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengunggah file ke storage"})
		return "", false
	}
	return objectName, true
}

func BackToSintesys(c *gin.Context) {
	_, mahasiswa, isError := getMahasiswa(c)

	if isError {
		RedirectSintesys(c)
		return
	}

	tagihanRepo := repositories.TagihanRepository{DB: database.DB}
	year, err := tagihanRepo.GetActiveFinanceYear()

	if err != nil {
		RedirectSintesys(c)
		return
	}

	if mahasiswa.UKT == "0" {
		hitAndBack(c, mahasiswa.MhswID, year.AcademicYear, mahasiswa.UKT)
		return
	}

	studentBills, err := tagihanRepo.GetStudentBills(mahasiswa.MhswID, year.AcademicYear)

	if err != nil {
		RedirectSintesys(c)
		return
	}

	isPaid := false
	for _, studentBill := range studentBills {
		if studentBill.PaidAmount > studentBill.Amount {
			isPaid = true
		}
	}
	if isPaid {
		hitAndBack(c, mahasiswa.MhswID, year.AcademicYear, mahasiswa.UKT)
		return
	}

	RedirectSintesys(c)
	return

}

func hitAndBack(c *gin.Context, studentId string, academicYear string, ukt string) {
	sintesysService := services.NewSintesys()
	sintesysService.SendCallback(studentId, academicYear, ukt)
	RedirectSintesys(c)
	return
}

func RedirectSintesys(c *gin.Context) {
	backUrl := os.Getenv("SINTESYS_URL")
	if backUrl == "" {
		backUrl = "http://sintesys.unsil.ac.id"
	}
	c.JSON(http.StatusOK, gin.H{
		"url": backUrl,
	})
	return
}
