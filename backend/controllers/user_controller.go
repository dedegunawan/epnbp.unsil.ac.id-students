package controllers

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/services"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
)

func getMahasiswa(c *gin.Context) (*models.User, *models.Mahasiswa, bool) {
	userRepo := repositories.UserRepository{DB: database.DB}
	ssoID := c.GetString("sso_id")
	email := c.GetString("email")
	name := c.GetString("name")

	utils.Log.Info("getMahasiswa", "sso_id", ssoID, "email", email, "name", name)

	var user *models.User
	var err error

	// Coba cari berdasarkan sso_id terlebih dahulu
	if ssoID != "" {
		user, err = userRepo.FindBySSOID(ssoID)
		if err == nil && user != nil {
			utils.Log.Info("User found by sso_id:", ssoID)
		} else {
			utils.Log.Info("User not found by sso_id:", ssoID, "error:", err)
		}
	}

	// Jika tidak ditemukan berdasarkan sso_id, coba cari berdasarkan email
	if user == nil && email != "" {
		utils.Log.Info("Trying to find user by email:", email)
		user, err = userRepo.FindByEmail(email)
		if err == nil && user != nil {
			utils.Log.Info("User found by email:", email, "user_id:", user.ID.String())
			// Update sso_id jika belum ada atau berbeda
			if user.SSOID == nil || (ssoID != "" && *user.SSOID != ssoID) {
				oldSSOID := "nil"
				if user.SSOID != nil {
					oldSSOID = *user.SSOID
				}
				utils.Log.Info("Updating user sso_id from", oldSSOID, "to", ssoID)
				user.SSOID = &ssoID
				if updateErr := userRepo.Update(user); updateErr != nil {
					utils.Log.Error("Failed to update user sso_id:", updateErr)
				} else {
					utils.Log.Info("Updated user sso_id successfully:", ssoID)
				}
			}
		} else {
			utils.Log.Info("User not found by email:", email, "error:", err)
		}
	} else if user == nil {
		utils.Log.Warn("Cannot search by email - email is empty or user already found")
	}

	// Jika user masih tidak ditemukan, buat user baru dari token claims
	if user == nil {
		if email == "" {
			utils.Log.Error("Cannot create user - email is empty", "sso_id", ssoID, "name", name)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required but not found in token"})
			return nil, nil, true
		}
		if ssoID == "" {
			utils.Log.Error("Cannot create user - sso_id is empty", "email", email, "name", name)
			c.JSON(http.StatusBadRequest, gin.H{"error": "SSO ID is required but not found in token"})
			return nil, nil, true
		}

		utils.Log.Info("Creating new user from token claims", "sso_id", ssoID, "email", email, "name", name)
		userService := services.UserService{Repo: &userRepo}
		user, err = userService.GetOrCreateByEmail(ssoID, email, name)
		if err != nil {
			utils.Log.Error("Failed to create user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
			return nil, nil, true
		}
		utils.Log.Info("User created successfully:", user.ID.String(), "email:", user.Email, "sso_id:", func() string {
			if user.SSOID != nil {
				return *user.SSOID
			}
			return "nil"
		}())
	}

	if user == nil {
		errorMsg := "User not found and could not be created"
		utils.Log.Error("User not found", "sso_id", ssoID, "email", email, "error", errorMsg)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return nil, nil, true
	}

	mahasiswaRepo := repositories.NewMahasiswaRepository(database.DB)
	userEmail := user.Email
	mahasiswa, _ := mahasiswaRepo.FindByEmailPattern(userEmail)

	mahasiswaID := "nil"
	if mahasiswa != nil {
		mahasiswaID = mahasiswa.MhswID
	}
	utils.Log.Info("mahasiswa found", "email", userEmail, "mahasiswa_id", mahasiswaID)

	return user, mahasiswa, false
}

func Me(c *gin.Context) {
	user, mahasiswa, mustreturn := getMahasiswa(c)
	if mustreturn {
		return
	}
	ssoID := c.GetString("sso_id")

	semester, err := semesterSaatIniMahasiswa(mahasiswa)
	if err != nil {
		utils.Log.Error(err.Error())
		semester = 0 // Jika ada error, set semester ke 0
	}

	c.JSON(200, gin.H{
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"sso_id":    ssoID,
		"is_active": user.IsActive,
		"mahasiswa": mahasiswa,
		"semester":  semester,
	})
}

func semesterSaatIniMahasiswa(mahasiswa *models.Mahasiswa) (int, error) {
	tagihanRepo := repositories.NewTagihanRepository(database.DB, database.DBPNBP)
	masterTagihanagihanRepo := repositories.MasterTagihanRepository{DB: database.DB}
	tagihanService := services.NewTagihanService(*tagihanRepo, masterTagihanagihanRepo)

	// Panggil repository untuk ambil FinanceYear aktif
	activeYear, err := tagihanRepo.GetActiveFinanceYearWithOverride(*mahasiswa)
	if err != nil {
		return 0, fmt.Errorf("Tahun aktif tidak ditemukan: %w", err)
	}

	// Prioritas 1: Ambil TahunID dari mahasiswa_masters di database PNBP
	TahunID := getTahunIDFromMahasiswaMasters(mahasiswa.MhswID)
	
	// Prioritas 2: Fallback ke ParseFullData (untuk kompatibilitas dengan data SIMAK/lama)
	if TahunID == "" {
		TahunID = getTahunIDFormParsed(mahasiswa)
	}
	
	if TahunID != "" {
		return tagihanService.HitungSemesterSaatIni(TahunID, activeYear.AcademicYear)
	}
	utils.Log.Info("TahunID tidak ditemukan pada data mahasiswa", "mahasiswa ", TahunID, " TahunID", mahasiswa.ParseFullData()["TahunID"])
	return 0, fmt.Errorf("TahunID tidak ditemukan pada data mahasiswa")
}

// getTahunIDFromMahasiswaMasters mengambil TahunID langsung dari mahasiswa_masters di database PNBP
func getTahunIDFromMahasiswaMasters(mhswID string) string {
	var mhswMaster models.MahasiswaMaster
	err := database.DBPNBP.Where("student_id = ?", mhswID).First(&mhswMaster).Error
	if err != nil {
		return ""
	}
	
	// TahunMasuk adalah int (contoh: 2023)
	// SemesterMasukID adalah uint (1 = Ganjil, 2 = Genap, atau sesuai enum)
	// Format TahunID: YYYYS (tahun + semester)
	// Jika SemesterMasukID tidak ada, default ke semester 1 (Ganjil)
	semesterMasuk := 1
	if mhswMaster.SemesterMasukID > 0 {
		semesterMasuk = int(mhswMaster.SemesterMasukID)
		// Pastikan semester hanya 1 atau 2
		if semesterMasuk > 2 {
			semesterMasuk = 1
		}
	}
	
	if mhswMaster.TahunMasuk > 0 {
		TahunID := fmt.Sprintf("%d%d", mhswMaster.TahunMasuk, semesterMasuk)
		utils.Log.Info("TahunID diambil dari mahasiswa_masters", "mhswID", mhswID, "TahunMasuk", mhswMaster.TahunMasuk, "SemesterMasukID", mhswMaster.SemesterMasukID, "TahunID", TahunID)
		return TahunID
	}
	
	return ""
}

func getTahunIDFormParsed(mahasiswa *models.Mahasiswa) string {
	data := mahasiswa.ParseFullData()
	
	// Coba ambil TahunID langsung
	tahunRaw, exists := data["TahunID"]
	if exists {
		var TahunID string
		switch v := tahunRaw.(type) {
		case string:
			TahunID = v
		case float64:
			TahunID = fmt.Sprintf("%.0f", v)
		case int:
			TahunID = strconv.Itoa(v)
		default:
			utils.Log.Info("TahunID ditemukan tapi tipe tidak dikenali", "value", tahunRaw)
			return ""
		}
		if TahunID != "" {
			return TahunID
		}
	}
	
	// Fallback: coba ambil dari TahunMasuk jika ada
	if tahunMasuk, ok := data["TahunMasuk"].(float64); ok {
		TahunID := fmt.Sprintf("%.0f1", tahunMasuk) // Default semester 1
		utils.Log.Info("TahunID dibuat dari TahunMasuk", "TahunMasuk", tahunMasuk, "TahunID", TahunID)
		return TahunID
	}
	
	utils.Log.Info("Field TahunID tidak ditemukan pada data mahasiswa", "data", data)
	return ""

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

	tagihanRepo := repositories.NewTagihanRepository(database.DB, database.DBPNBP)

	// Panggil repository untuk ambil FinanceYear aktif
	activeYear, err := tagihanRepo.GetActiveFinanceYearWithOverride(*mahasiswa)
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
	tagihanRepo := repositories.NewTagihanRepository(database.DB, database.DBPNBP)

	// Panggil repository untuk ambil FinanceYear aktif
	activeYear, err := tagihanRepo.GetActiveFinanceYearWithOverride(*mahasiswa)
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

	var statusAkademikId int
	if v, ok := full_data["StatusAkademikID"].(float64); ok {
		statusAkademikId = int(v)
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

	tagihanRepo := repositories.NewTagihanRepository(database.DB, database.DBPNBP)

	// Panggil repository untuk ambil FinanceYear aktif
	activeYear, err := tagihanRepo.GetActiveFinanceYearWithOverride(*mahasiswa)
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

	masterTagihanagihanRepo := repositories.MasterTagihanRepository{DB: database.DBPNBP}
	tagihanService := services.NewTagihanService(*tagihanRepo, masterTagihanagihanRepo)

	if len(tagihan) == 0 {
		utils.Log.Info("Memulai pembuatan tagihan baru", map[string]interface{}{
			"mhswID":      mahasiswa.MhswID,
			"nama":        mahasiswa.Nama,
			"BIPOTID":     mahasiswa.BIPOTID,
			"UKT":         mahasiswa.UKT,
			"academicYear": activeYear.AcademicYear,
		})
		if err := tagihanService.CreateNewTagihan(mahasiswa, activeYear); err != nil {
			utils.Log.Error("Gagal membuat tagihan", map[string]interface{}{
				"mhswID":      mahasiswa.MhswID,
				"nama":        mahasiswa.Nama,
				"BIPOTID":     mahasiswa.BIPOTID,
				"UKT":         mahasiswa.UKT,
				"academicYear": activeYear.AcademicYear,
				"error":       err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Gagal membuat tagihan",
				"message": err.Error(),
				"details": map[string]interface{}{
					"mhswID":      mahasiswa.MhswID,
					"BIPOTID":     mahasiswa.BIPOTID,
					"UKT":         mahasiswa.UKT,
					"academicYear": activeYear.AcademicYear,
				},
			})
			return
		}
		utils.Log.Info("Tagihan berhasil dibuat", map[string]interface{}{
			"mhswID":      mahasiswa.MhswID,
			"academicYear": activeYear.AcademicYear,
		})
	}

	masihAdaKurangNominal := false

	mangajukanCicilan := tagihanService.CekCicilanMahasiswa(mahasiswa, activeYear)
	tidakMengajukanCicilan := !mangajukanCicilan
	mengajukanPenangguhan := tagihanService.CekPenangguhanMahasiswa(mahasiswa, activeYear)
	tidakMengajukanPenangguhan := !mengajukanPenangguhan
	mendapatBeasiswa := tagihanService.CekBeasiswaMahasiswa(mahasiswa, activeYear)
	tidakMendapatBeasiswa := !mendapatBeasiswa
	//punyaDeposit := tagihanService.CekDepositMahasiswa(mahasiswa, activeYear)
	//tidakPunyaDeposit := !punyaDeposit

	nominalDitagihaLebihKecilSeharusnya, tagihanSeharusnya, totalTagihanDibayar := tagihanService.IsNominalDibayarLebihKecilSeharusnya(mahasiswa, activeYear)

	nominalKurangBayar := tagihanSeharusnya - totalTagihanDibayar

	if tidakMengajukanCicilan && tidakMengajukanPenangguhan && tidakMendapatBeasiswa && nominalDitagihaLebihKecilSeharusnya {
		masihAdaKurangNominal = true
	}

	utils.Log.Info("Haruskah buat tagihan baru? ", len(tagihan) > 0 && masihAdaKurangNominal, len(tagihan) > 0, masihAdaKurangNominal, tidakMengajukanCicilan, tidakMengajukanPenangguhan, tidakMendapatBeasiswa, nominalDitagihaLebihKecilSeharusnya)

	if len(tagihan) > 0 && masihAdaKurangNominal {
		if err := tagihanService.CreateNewTagihanSekurangnya(mahasiswa, activeYear, nominalKurangBayar); err != nil {
			utils.Log.Info("Gagal membuat tagihan", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat tagihan"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GenerateCurrentBillPascasarjana(c *gin.Context, mahasiswa models.Mahasiswa) {
	utils.Log.Info("GenerateCurrentBillPascasarjana")
	tagihanRepo := repositories.NewTagihanRepository(database.DB, database.DBPNBP)

	// Panggil repository untuk ambil FinanceYear aktif
	activeYear, err := tagihanRepo.GetActiveFinanceYearWithOverride(mahasiswa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tahun aktif tidak ditemukan"})
		return
	}

	// Panggil repository untuk ambil FinanceYear aktif
	if !GetIsMahasiswaAktifFromFullData(mahasiswa) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Pembuatan tagihan baru untuk tahun aktif, hanya diperboleh untuk mahasiswa aktif"})
		return
	}

	masterTagihanagihanRepo := repositories.MasterTagihanRepository{DB: database.DB}
	tagihanService := services.NewTagihanService(*tagihanRepo, masterTagihanagihanRepo)

	if err := tagihanService.CreateNewTagihanPasca(&mahasiswa, activeYear); err != nil {
		utils.Log.Info("Gagal membuat tagihan", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat tagihan"})
		return
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

	tagihanRepo := repositories.NewTagihanRepository(database.DB, database.DBPNBP)

	// Panggil repository untuk ambil FinanceYear aktif
	activeYear, err := tagihanRepo.GetActiveFinanceYearWithOverride(*mahasiswa)
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
	tagihanRepo := repositories.NewTagihanRepository(database.DB, database.DBPNBP)
	studentBill, err := tagihanRepo.FindStudentBillByID(studentBillID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tagihan tidak ditemukan"})
		return
	}

	fileURL, ok := handleUpload(c, "file")
	if !ok {
		return
	}

	masterTagihanagihanRepo := repositories.MasterTagihanRepository{DB: database.DB}

	// Simpan ke database (opsional, sesuaikan dengan struktur Anda)
	paymentConfirmation, err := services.NewTagihanService(*tagihanRepo, masterTagihanagihanRepo).SavePaymentConfirmation(*studentBill, vaNumber, paymentDate, fileURL)
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

	tagihanRepo := repositories.NewTagihanRepository(database.DB, database.DBPNBP)
	year, err := tagihanRepo.GetActiveFinanceYearWithOverride(*mahasiswa)

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
