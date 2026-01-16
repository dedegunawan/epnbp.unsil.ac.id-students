package controllers

import (
	"net/http"
	"os"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/services"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
)

// GetStudentBillStatusNew GET /api/v1/student-bill-new
// Menampilkan tagihan mahasiswa dari cicilan atau registrasi (tanpa student_bill)
func GetStudentBillStatusNew(c *gin.Context) {
	mhswMaster, mustreturn := getMahasiswa(c)
	if mustreturn {
		return
	}

	if mhswMaster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data mahasiswa tidak ditemukan"})
		return
	}

	tagihanRepo := repositories.NewTagihanRepository(database.DBPNBP, database.DBPNBP)

	// Ambil FinanceYear aktif (tidak perlu override karena tidak ada mahasiswa lokal)
	activeYear, err := tagihanRepo.GetActiveFinanceYear()
	if err != nil {
		utils.Log.Error("Gagal mengambil finance year aktif", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tahun aktif tidak ditemukan"})
		return
	}

	// Buat dummy mahasiswa dari mhswMaster untuk service
	dummyMahasiswa := &models.Mahasiswa{
		MhswID: mhswMaster.StudentID,
		Nama:   mhswMaster.NamaLengkap,
	}

	// Buat service baru
	tagihanNewService := services.NewTagihanNewService(*tagihanRepo)

	// Ambil tagihan yang harus dibayar dari cicilan atau registrasi
	tagihanList, err := tagihanNewService.GetTagihanMahasiswa(dummyMahasiswa, activeYear)
	if err != nil {
		utils.Log.Error("Gagal mengambil tagihan", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengambil tagihan",
			"message": err.Error(),
		})
		return
	}

	// Ambil riwayat pembayaran dari registrasi_mahasiswa dan detail_cicilan
	historyList, err := tagihanNewService.GetHistoryTagihanMahasiswa(dummyMahasiswa, activeYear)
	if err != nil {
		utils.Log.Error("Gagal mengambil riwayat pembayaran", "error", err.Error())
		// Jangan return error, cukup log dan gunakan history kosong
		historyList = []models.TagihanResponse{}
	}

	// Pisahkan tagihan harus dibayar
	var tagihanHarusDibayar []models.TagihanResponse
	allPaid := true

	for _, t := range tagihanList {
		if t.RemainingAmount > 0 {
			tagihanHarusDibayar = append(tagihanHarusDibayar, t)
			allPaid = false
		}
	}

	isGenerated := len(tagihanList) > 0 || len(historyList) > 0
	if !isGenerated {
		allPaid = false
	}

	response := models.TagihanListResponse{
		Tahun:               *activeYear,
		IsPaid:              allPaid,
		IsGenerated:         isGenerated,
		TagihanHarusDibayar: tagihanHarusDibayar,
		HistoryTagihan:      historyList,
	}

	c.JSON(http.StatusOK, response)
}

// GenerateUrlPembayaranNew GET /api/v1/generate-payment-new
// Generate URL pembayaran untuk tagihan dari cicilan atau registrasi
// Query params:
//   - detail_cicilan_id: ID dari detail_cicilan (untuk tagihan cicilan)
//   - registrasi_mahasiswa_id: ID dari registrasi_mahasiswa (untuk tagihan registrasi)
func GenerateUrlPembayaranNew(c *gin.Context) {
	detailCicilanID := c.Query("detail_cicilan_id")
	registrasiMahasiswaID := c.Query("registrasi_mahasiswa_id")

	// Ambil EPNBP_URL dari environment variable
	epnbpURL := os.Getenv("EPNBP_URL")
	if epnbpURL == "" {
		epnbpURL = "https://epnbp.unsil.ac.id" // Default URL
	}

	var redirectURL string

	// Cek apakah ada detail_cicilan_id
	if detailCicilanID != "" {
		// URL: EPNBP_URL + "/api/generate-va?detail_cicilan_id=" + id
		redirectURL = epnbpURL + "/api/generate-va?detail_cicilan_id=" + detailCicilanID
		utils.Log.Info("Generate payment URL untuk cicilan", map[string]interface{}{
			"detail_cicilan_id": detailCicilanID,
			"redirect_url":      redirectURL,
		})
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// Cek apakah ada registrasi_mahasiswa_id
	if registrasiMahasiswaID != "" {
		// URL: EPNBP_URL + "/api/generate-va?registrasi_mahasiswa_id=" + id
		redirectURL = epnbpURL + "/api/generate-va?registrasi_mahasiswa_id=" + registrasiMahasiswaID
		utils.Log.Info("Generate payment URL untuk registrasi", map[string]interface{}{
			"registrasi_mahasiswa_id": registrasiMahasiswaID,
			"redirect_url":             redirectURL,
		})
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// Jika tidak ada parameter yang valid
	utils.Log.Warn("Generate payment URL: parameter tidak valid", map[string]interface{}{
		"detail_cicilan_id":       detailCicilanID,
		"registrasi_mahasiswa_id": registrasiMahasiswaID,
	})
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "Parameter tidak valid",
		"message": "Harus menyertakan detail_cicilan_id atau registrasi_mahasiswa_id",
	})
}
