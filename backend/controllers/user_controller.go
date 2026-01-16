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

	"github.com/dedegunawan/backend-ujian-telp-v5/config"
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/services"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
)

func getMahasiswa(c *gin.Context) (*models.MahasiswaMaster, bool) {
	// Ambil email dari context (token SSO) - tidak perlu query ke tabel users
	email := c.GetString("email")
	ssoID := c.GetString("sso_id")
	name := c.GetString("name")

	utils.Log.Info("getMahasiswa", map[string]interface{}{
		"sso_id": ssoID,
		"email":  email,
		"name":   name,
	})

	// Validasi email tidak kosong
	if email == "" {
		utils.Log.Error("Email kosong dari context", map[string]interface{}{
			"sso_id": ssoID,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email tidak ditemukan dalam token"})
		return nil, true
	}

	// Validasi email suffix
	if !config.ValidateEmailSuffix(email) {
		utils.Log.Error("Email suffix tidak valid", map[string]interface{}{
			"email":           email,
			"required_suffix": config.GetEmailSuffix(),
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Email harus menggunakan domain %s", config.GetEmailSuffix())})
		return nil, true
	}

	// Ambil studentID dari email
	studentID := utils.GetEmailPrefix(email)
	if studentID == "" {
		utils.Log.Error("StudentID kosong dari email", map[string]interface{}{
			"email": email,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak dapat mengambil NPM dari email"})
		return nil, true
	}

	// Ambil data langsung dari mahasiswa_masters (tidak perlu query ke users)
	var mhswMaster models.MahasiswaMaster
	err := database.DBPNBP.Preload("MasterTagihan").Where("student_id = ?", studentID).First(&mhswMaster).Error
	if err != nil {
		utils.Log.Error("Gagal mengambil data dari mahasiswa_masters", map[string]interface{}{
			"studentID": studentID,
			"email":     email,
			"error":     err.Error(),
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "Data mahasiswa tidak ditemukan"})
		return nil, true
	}

	utils.Log.Info("Mahasiswa ditemukan dari mahasiswa_masters", map[string]interface{}{
		"studentID": studentID,
		"email":     email,
		"nama":      mhswMaster.NamaLengkap,
	})

	return &mhswMaster, false
}

func Me(c *gin.Context) {
	utils.Log.Info("Endpoint /me dipanggil")

	// Ambil data langsung dari mahasiswa_masters (tidak perlu query ke users)
	mhswMaster, mustreturn := getMahasiswa(c)
	if mustreturn {
		utils.Log.Warn("Endpoint /me: getMahasiswa mengembalikan mustreturn=true")
		return
	}

	if mhswMaster == nil {
		utils.Log.Warn("Endpoint /me: Mahasiswa tidak ditemukan")
		c.JSON(http.StatusNotFound, gin.H{"error": "Data mahasiswa tidak ditemukan"})
		return
	}

	// Ambil prodi dan fakultas dari database PNBP
	var prodiPnbp models.ProdiPnbp
	var fakultasPnbp models.FakultasPnbp
	if mhswMaster.ProdiID > 0 {
		database.DBPNBP.Where("id = ?", mhswMaster.ProdiID).First(&prodiPnbp)
		if prodiPnbp.ID > 0 && prodiPnbp.FakultasID > 0 {
			database.DBPNBP.Where("id = ?", prodiPnbp.FakultasID).First(&fakultasPnbp)
		}
	}

	// Ambil status mahasiswa dari mahasiswa_masters via status_akademiks
	var statusMahasiswa string = "Non-Aktif"
	var statusKode string = "N"
	if mhswMaster.StatusAkademikID > 0 {
		var statusAkademik models.StatusAkademik
		errStatus := database.DBPNBP.Where("id = ?", mhswMaster.StatusAkademikID).First(&statusAkademik).Error
		if errStatus == nil && statusAkademik.Kode != "" {
			statusKode = statusAkademik.Kode
			statusMahasiswa = statusAkademik.Nama
			utils.Log.Info("Endpoint /me: Status mahasiswa diambil dari status_akademiks", map[string]interface{}{
				"mhswID":           mhswMaster.StudentID,
				"StatusAkademikID": mhswMaster.StatusAkademikID,
				"Kode":             statusAkademik.Kode,
				"Nama":             statusAkademik.Nama,
			})
		} else {
			utils.Log.Warn("Endpoint /me: Status akademik tidak ditemukan", map[string]interface{}{
				"mhswID":           mhswMaster.StudentID,
				"StatusAkademikID": mhswMaster.StatusAkademikID,
				"error":            errStatus,
			})
		}
	} else {
		utils.Log.Warn("Endpoint /me: StatusAkademikID = 0", map[string]interface{}{
			"mhswID": mhswMaster.StudentID,
		})
	}

	// Hitung semester dari mahasiswa_masters
	semester := 0
	semester, err := semesterSaatIniMahasiswaFromMaster(mhswMaster, nil)
	if err != nil {
		utils.Log.Error("Endpoint /me: Gagal menghitung semester", map[string]interface{}{
			"mhswID": mhswMaster.StudentID,
			"error":  err.Error(),
		})
		semester = 0
	}

	// Format UKT kelompok (decimal ke int, lalu ke string)
	kelompokUKT := strconv.Itoa(int(mhswMaster.UKT))

	// Ambil UKTNominal dari detail_tagihan
	nominalUKTFromDetail := int64(0)
	if mhswMaster.MasterTagihanID == 0 {
		utils.Log.Warn("Endpoint /me: MasterTagihanID adalah 0, tidak bisa mengambil UKTNominal", map[string]interface{}{
			"mhswID": mhswMaster.StudentID,
			"UKT":    mhswMaster.UKT,
		})
	} else if mhswMaster.UKT == 0 {
		utils.Log.Warn("Endpoint /me: UKT adalah 0, tidak bisa mengambil UKTNominal", map[string]interface{}{
			"mhswID":          mhswMaster.StudentID,
			"masterTagihanID": mhswMaster.MasterTagihanID,
		})
	} else {
		// UKT dan MasterTagihanID valid, ambil dari detail_tagihan
		var detailTagihan models.DetailTagihan
		UKTStr := strconv.Itoa(int(mhswMaster.UKT))

		// Coba query dengan format int sebagai string
		errDetail := database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, UKTStr).
			First(&detailTagihan).Error

		if errDetail == nil {
			nominalUKTFromDetail = detailTagihan.Nominal
		} else {
			// Fallback: coba format float dengan 2 desimal
			UKTFloat := fmt.Sprintf("%.2f", mhswMaster.UKT)
			errDetail = database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, UKTFloat).
				First(&detailTagihan).Error
			if errDetail == nil {
				nominalUKTFromDetail = detailTagihan.Nominal
			} else {
				// Fallback: coba tanpa desimal
				UKTNoDecimal := fmt.Sprintf("%.0f", mhswMaster.UKT)
				errDetail = database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, UKTNoDecimal).
					First(&detailTagihan).Error
				if errDetail == nil {
					nominalUKTFromDetail = detailTagihan.Nominal
				}
			}
		}
	}

	// Buat parsedData
	parsedData := make(map[string]interface{})
	parsedData["TahunMasuk"] = mhswMaster.TahunMasuk
	parsedData["angkatan"] = strconv.Itoa(mhswMaster.TahunMasuk)
	parsedData["UKT"] = int(mhswMaster.UKT)
	parsedData["UKTNominal"] = nominalUKTFromDetail
	parsedData["master_tagihan_id"] = mhswMaster.MasterTagihanID
	parsedData["StatusMhswID"] = statusKode // Status dari mahasiswa_masters
	parsedData["StatusAkademikID"] = mhswMaster.StatusAkademikID
	parsedData["StatusNama"] = statusMahasiswa

	// Buat response mahasiswa dari mahasiswa_masters
	mahasiswaResponse := gin.H{
		"mhsw_id":     mhswMaster.StudentID,
		"nama":        mhswMaster.NamaLengkap,
		"kel_ukt":     kelompokUKT,
		"bipot_id":    fmt.Sprintf("%d", mhswMaster.MasterTagihanID),
		"email":       mhswMaster.Email,
		"tahun_masuk": mhswMaster.TahunMasuk,
		"status":      statusMahasiswa, // Status dari mahasiswa_masters
		"status_kode": statusKode,      // Kode status (A, N, dll)
		"parsed":      parsedData,
		"prodi": gin.H{
			"kode_prodi":  prodiPnbp.KodeProdi,
			"nama_prodi":  prodiPnbp.NamaProdi,
			"fakultas_id": prodiPnbp.FakultasID,
			"fakultas": gin.H{
				"kode_fakultas": fakultasPnbp.KodeFakultas,
				"nama_fakultas": fakultasPnbp.NamaFakultas,
			},
		},
	}

	email := c.GetString("email")
	ssoID := c.GetString("sso_id")
	name := c.GetString("name")

	response := gin.H{
		"name":      name,
		"email":     email,
		"sso_id":    ssoID,
		"mahasiswa": mahasiswaResponse,
		"semester":  semester,
	}

	utils.Log.Info("Endpoint /me: Response berhasil", map[string]interface{}{
		"mhswID":   mhswMaster.StudentID,
		"semester": semester,
	})

	c.JSON(200, response)
}

func semesterSaatIniMahasiswa(mahasiswa *models.Mahasiswa) (int, error) {
	// Validasi mahasiswa dan MhswID tidak kosong
	if mahasiswa == nil {
		return 0, fmt.Errorf("mahasiswa tidak ditemukan")
	}
	if mahasiswa.MhswID == "" {
		return 0, fmt.Errorf("MhswID kosong")
	}

	utils.Log.Info("semesterSaatIniMahasiswa: Memulai perhitungan semester", map[string]interface{}{
		"mhswID": mahasiswa.MhswID,
		"nama":   mahasiswa.Nama,
	})

	tagihanRepo := repositories.NewTagihanRepository(database.DBPNBP, database.DBPNBP)
	masterTagihanagihanRepo := repositories.MasterTagihanRepository{DB: database.DBPNBP}
	tagihanService := services.NewTagihanService(*tagihanRepo, masterTagihanagihanRepo)

	// Panggil repository untuk ambil FinanceYear aktif (berisi budget_periods.kode)
	activeYear, err := tagihanRepo.GetActiveFinanceYearWithOverride(*mahasiswa)
	if err != nil {
		utils.Log.Error("semesterSaatIniMahasiswa: Gagal ambil FinanceYear aktif", map[string]interface{}{
			"mhswID": mahasiswa.MhswID,
			"error":  err.Error(),
		})
		return 0, fmt.Errorf("Tahun aktif tidak ditemukan: %w", err)
	}

	utils.Log.Info("semesterSaatIniMahasiswa: FinanceYear aktif ditemukan", map[string]interface{}{
		"mhswID":       mahasiswa.MhswID,
		"academicYear": activeYear.AcademicYear,
		"code":         activeYear.Code,
		"description":  activeYear.Description,
	})

	// Ambil SemesterMasukID dari mahasiswa_masters di database PNBP
	// SemesterMasukID adalah referensi ke budget_periods.id, bukan semester masuk (1/2)
	var mhswMaster models.MahasiswaMaster
	errMhswMaster := database.DBPNBP.Where("student_id = ?", mahasiswa.MhswID).First(&mhswMaster).Error

	var tahunIDAwal string

	if errMhswMaster == nil && mhswMaster.SemesterMasukID > 0 {
		// SemesterMasukID adalah referensi ke budget_periods.id
		// Ambil budget_periods.kode sebagai tahunIDAwal
		var budgetPeriod models.BudgetPeriod
		errBudgetPeriod := database.DBPNBP.Where("id = ?", mhswMaster.SemesterMasukID).First(&budgetPeriod).Error
		if errBudgetPeriod == nil && budgetPeriod.Kode != "" {
			tahunIDAwal = budgetPeriod.Kode
			utils.Log.Info("semesterSaatIniMahasiswa: TahunID awal diambil dari budget_periods berdasarkan SemesterMasukID", map[string]interface{}{
				"mhswID":           mahasiswa.MhswID,
				"SemesterMasukID":  mhswMaster.SemesterMasukID,
				"budgetPeriodID":   budgetPeriod.ID,
				"budgetPeriodKode": budgetPeriod.Kode,
				"tahunIDAwal":      tahunIDAwal,
			})
		} else {
			utils.Log.Warn("semesterSaatIniMahasiswa: Budget period tidak ditemukan berdasarkan SemesterMasukID, fallback ke TahunMasuk", map[string]interface{}{
				"mhswID":          mahasiswa.MhswID,
				"SemesterMasukID": mhswMaster.SemesterMasukID,
				"error":           errBudgetPeriod,
			})
			// Fallback: gunakan TahunMasuk dengan default semester 1
			if mhswMaster.TahunMasuk > 0 {
				tahunIDAwal = fmt.Sprintf("%d1", mhswMaster.TahunMasuk)
				utils.Log.Info("semesterSaatIniMahasiswa: TahunID awal dibuat dari TahunMasuk (fallback)", map[string]interface{}{
					"mhswID":      mahasiswa.MhswID,
					"TahunMasuk":  mhswMaster.TahunMasuk,
					"tahunIDAwal": tahunIDAwal,
				})
			} else {
				utils.Log.Error("semesterSaatIniMahasiswa: TahunMasuk juga tidak ditemukan", map[string]interface{}{
					"mhswID": mahasiswa.MhswID,
				})
				return 0, fmt.Errorf("tahun masuk tidak ditemukan untuk mahasiswa %s", mahasiswa.MhswID)
			}
		}
	} else {
		utils.Log.Warn("semesterSaatIniMahasiswa: Mahasiswa tidak ditemukan di mahasiswa_masters atau SemesterMasukID = 0, fallback ke FullData", map[string]interface{}{
			"mhswID": mahasiswa.MhswID,
			"error":  errMhswMaster,
		})
		// Fallback: ambil dari FullData
		if tahunIDData, ok := mahasiswa.ParseFullData()["TahunID"].(string); ok && tahunIDData != "" {
			tahunIDAwal = tahunIDData
			utils.Log.Info("semesterSaatIniMahasiswa: TahunID awal diambil dari FullData", map[string]interface{}{
				"mhswID":      mahasiswa.MhswID,
				"tahunIDAwal": tahunIDAwal,
			})
		} else if tahunMasukData, ok := mahasiswa.ParseFullData()["TahunMasuk"].(float64); ok {
			// Fallback: buat dari TahunMasuk dengan default semester 1
			tahunIDAwal = fmt.Sprintf("%.0f1", tahunMasukData)
			utils.Log.Info("semesterSaatIniMahasiswa: TahunID awal dibuat dari TahunMasuk di FullData", map[string]interface{}{
				"mhswID":      mahasiswa.MhswID,
				"TahunMasuk":  tahunMasukData,
				"tahunIDAwal": tahunIDAwal,
			})
		} else {
			utils.Log.Error("semesterSaatIniMahasiswa: TahunID atau TahunMasuk tidak ditemukan di FullData", map[string]interface{}{
				"mhswID":   mahasiswa.MhswID,
				"fullData": mahasiswa.ParseFullData(),
			})
			return 0, fmt.Errorf("tahun masuk tidak ditemukan untuk mahasiswa %s", mahasiswa.MhswID)
		}
	}
	utils.Log.Info("semesterSaatIniMahasiswa: Memanggil HitungSemesterSaatIni", map[string]interface{}{
		"mhswID":       mahasiswa.MhswID,
		"tahunIDAwal":  tahunIDAwal,
		"academicYear": activeYear.AcademicYear,
	})

	semester, err := tagihanService.HitungSemesterSaatIni(tahunIDAwal, activeYear.AcademicYear)
	if err != nil {
		utils.Log.Error("semesterSaatIniMahasiswa: Gagal hitung semester", map[string]interface{}{
			"mhswID":       mahasiswa.MhswID,
			"tahunIDAwal":  tahunIDAwal,
			"academicYear": activeYear.AcademicYear,
			"error":        err.Error(),
		})
		return 0, err
	}

	utils.Log.Info("semesterSaatIniMahasiswa: Semester berhasil dihitung", map[string]interface{}{
		"mhswID":       mahasiswa.MhswID,
		"tahunIDAwal":  tahunIDAwal,
		"academicYear": activeYear.AcademicYear,
		"semester":     semester,
	})

	return semester, nil
}

// semesterSaatIniMahasiswaFromMaster menghitung semester langsung dari mahasiswa_masters di database PNBP
func semesterSaatIniMahasiswaFromMaster(mhswMaster *models.MahasiswaMaster, mahasiswa *models.Mahasiswa) (int, error) {
	if mhswMaster == nil {
		return 0, fmt.Errorf("mahasiswa master tidak ditemukan")
	}
	if mhswMaster.StudentID == "" {
		return 0, fmt.Errorf("StudentID kosong")
	}

	// Buat dummy mahasiswa untuk GetActiveFinanceYearWithOverride jika mahasiswa nil
	var dummyMahasiswa models.Mahasiswa
	if mahasiswa == nil {
		dummyMahasiswa.MhswID = mhswMaster.StudentID
		mahasiswa = &dummyMahasiswa
	}

	utils.Log.Info("semesterSaatIniMahasiswaFromMaster: Memulai perhitungan semester", map[string]interface{}{
		"mhswID": mhswMaster.StudentID,
		"nama":   mhswMaster.NamaLengkap,
	})

	tagihanRepo := repositories.NewTagihanRepository(database.DBPNBP, database.DBPNBP)
	masterTagihanRepo := repositories.MasterTagihanRepository{DB: database.DBPNBP}
	tagihanService := services.NewTagihanService(*tagihanRepo, masterTagihanRepo)

	// Ambil FinanceYear aktif langsung dari budget_periods (tidak perlu override karena tidak ada mahasiswa lokal)
	activeYear, err := tagihanRepo.GetActiveFinanceYear()
	if err != nil {
		utils.Log.Error("semesterSaatIniMahasiswaFromMaster: Gagal ambil FinanceYear aktif", map[string]interface{}{
			"mhswID": mhswMaster.StudentID,
			"error":  err.Error(),
		})
		return 0, fmt.Errorf("Tahun aktif tidak ditemukan: %w", err)
	}

	utils.Log.Info("semesterSaatIniMahasiswaFromMaster: FinanceYear aktif ditemukan", map[string]interface{}{
		"mhswID":       mhswMaster.StudentID,
		"academicYear": activeYear.AcademicYear,
		"code":         activeYear.Code,
		"description":  activeYear.Description,
	})

	var tahunIDAwal string

	if mhswMaster.SemesterMasukID > 0 {
		// SemesterMasukID adalah referensi ke budget_periods.id
		// Ambil budget_periods.kode sebagai tahunIDAwal
		var budgetPeriod models.BudgetPeriod
		errBudgetPeriod := database.DBPNBP.Where("id = ?", mhswMaster.SemesterMasukID).First(&budgetPeriod).Error
		if errBudgetPeriod == nil && budgetPeriod.Kode != "" {
			tahunIDAwal = budgetPeriod.Kode
			utils.Log.Info("semesterSaatIniMahasiswaFromMaster: TahunID awal diambil dari budget_periods berdasarkan SemesterMasukID", map[string]interface{}{
				"mhswID":           mhswMaster.StudentID,
				"SemesterMasukID":  mhswMaster.SemesterMasukID,
				"budgetPeriodID":   budgetPeriod.ID,
				"budgetPeriodKode": budgetPeriod.Kode,
				"tahunIDAwal":      tahunIDAwal,
			})
		} else {
			utils.Log.Warn("semesterSaatIniMahasiswaFromMaster: Budget period tidak ditemukan berdasarkan SemesterMasukID, fallback ke TahunMasuk", map[string]interface{}{
				"mhswID":          mhswMaster.StudentID,
				"SemesterMasukID": mhswMaster.SemesterMasukID,
				"error":           errBudgetPeriod,
			})
			// Fallback: gunakan TahunMasuk dengan default semester 1
			if mhswMaster.TahunMasuk > 0 {
				tahunIDAwal = fmt.Sprintf("%d1", mhswMaster.TahunMasuk)
				utils.Log.Info("semesterSaatIniMahasiswaFromMaster: TahunID awal dibuat dari TahunMasuk (fallback)", map[string]interface{}{
					"mhswID":      mhswMaster.StudentID,
					"TahunMasuk":  mhswMaster.TahunMasuk,
					"tahunIDAwal": tahunIDAwal,
				})
			} else {
				utils.Log.Error("semesterSaatIniMahasiswaFromMaster: TahunMasuk juga tidak ditemukan", map[string]interface{}{
					"mhswID": mhswMaster.StudentID,
				})
				return 0, fmt.Errorf("tahun masuk tidak ditemukan untuk mahasiswa %s", mhswMaster.StudentID)
			}
		}
	} else {
		// Fallback: gunakan TahunMasuk dengan default semester 1
		if mhswMaster.TahunMasuk > 0 {
			tahunIDAwal = fmt.Sprintf("%d1", mhswMaster.TahunMasuk)
			utils.Log.Info("semesterSaatIniMahasiswaFromMaster: TahunID awal dibuat dari TahunMasuk (SemesterMasukID tidak ada)", map[string]interface{}{
				"mhswID":      mhswMaster.StudentID,
				"TahunMasuk":  mhswMaster.TahunMasuk,
				"tahunIDAwal": tahunIDAwal,
			})
		} else {
			utils.Log.Error("semesterSaatIniMahasiswaFromMaster: TahunMasuk tidak ditemukan", map[string]interface{}{
				"mhswID": mhswMaster.StudentID,
			})
			return 0, fmt.Errorf("tahun masuk tidak ditemukan untuk mahasiswa %s", mhswMaster.StudentID)
		}
	}

	utils.Log.Info("semesterSaatIniMahasiswaFromMaster: Memanggil HitungSemesterSaatIni", map[string]interface{}{
		"mhswID":       mhswMaster.StudentID,
		"tahunIDAwal":  tahunIDAwal,
		"academicYear": activeYear.AcademicYear,
	})

	semester, err := tagihanService.HitungSemesterSaatIni(tahunIDAwal, activeYear.AcademicYear)
	if err != nil {
		utils.Log.Error("semesterSaatIniMahasiswaFromMaster: Gagal hitung semester", map[string]interface{}{
			"mhswID":       mhswMaster.StudentID,
			"tahunIDAwal":  tahunIDAwal,
			"academicYear": activeYear.AcademicYear,
			"error":        err.Error(),
		})
		return 0, err
	}

	utils.Log.Info("semesterSaatIniMahasiswaFromMaster: Semester berhasil dihitung", map[string]interface{}{
		"mhswID":       mhswMaster.StudentID,
		"tahunIDAwal":  tahunIDAwal,
		"academicYear": activeYear.AcademicYear,
		"semester":     semester,
	})

	return semester, nil
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
	// Tidak menggunakan endpoint ini lagi - gunakan GetStudentBillStatusNew
	c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint deprecated, gunakan /student-bill-new"})
	return
}

// POST /student-bill
func RegenerateCurrentBill(c *gin.Context) {
	// Endpoint ini deprecated - tidak perlu regenerate tagihan lagi
	c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint deprecated, tagihan langsung dari cicilan/registrasi"})
	return
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
	// Endpoint ini deprecated - tidak perlu generate tagihan lagi
	c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint deprecated, tagihan langsung dari cicilan/registrasi"})
	return
}

func GenerateCurrentBillPascasarjana(c *gin.Context, mahasiswa models.Mahasiswa) {
	utils.Log.Info("GenerateCurrentBillPascasarjana")
	tagihanRepo := repositories.NewTagihanRepository(database.DBPNBP, database.DBPNBP)

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

	masterTagihanagihanRepo := repositories.MasterTagihanRepository{DB: database.DBPNBP}
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

	epnbpRepo := repositories.NewEpnbpRepository(database.DBPNBP)

	payUrl, _ := epnbpRepo.FindNotExpiredByStudentBill(studentBillID)
	if payUrl != nil && payUrl.PayUrl != "" {
		c.JSON(http.StatusOK, payUrl)
		return
	}

	// Endpoint ini deprecated - gunakan GenerateUrlPembayaranNew
	c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint deprecated, gunakan /generate-payment-new"})
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
	tagihanRepo := repositories.NewTagihanRepository(database.DBPNBP, database.DBPNBP)
	studentBill, err := tagihanRepo.FindStudentBillByID(studentBillID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tagihan tidak ditemukan"})
		return
	}

	fileURL, ok := handleUpload(c, "file")
	if !ok {
		return
	}

	masterTagihanagihanRepo := repositories.MasterTagihanRepository{DB: database.DBPNBP}

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
	mhswMaster, isError := getMahasiswa(c)

	if isError || mhswMaster == nil {
		RedirectSintesys(c)
		return
	}

	tagihanRepo := repositories.NewTagihanRepository(database.DBPNBP, database.DBPNBP)
	year, err := tagihanRepo.GetActiveFinanceYear()

	if err != nil {
		RedirectSintesys(c)
		return
	}

	UKTStr := strconv.Itoa(int(mhswMaster.UKT))
	if UKTStr == "0" {
		hitAndBack(c, mhswMaster.StudentID, year.AcademicYear, UKTStr)
		return
	}

	// Cek tagihan dari cicilan atau registrasi (tidak perlu cek student_bill)
	// Langsung redirect ke sintesys
	hitAndBack(c, mhswMaster.StudentID, year.AcademicYear, UKTStr)
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
