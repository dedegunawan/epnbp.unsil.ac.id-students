package controllers

import (
	"bytes"
	_ "embed"
	"fmt"
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
	"github.com/jung-kurt/gofpdf/v2"
)

//go:embed logo.png
var embeddedLogo []byte

// getVirtualAccount mengambil virtual account dari database berdasarkan payment
// Langsung dari invoice_relations ke virtual_accounts via virtual_account_id
func getVirtualAccount(payment *models.TagihanResponse) string {
	var virtualAccount string

	if payment.Source == "cicilan" && payment.DetailCicilanID != nil {
		// Query langsung dari invoice_relations ke virtual_accounts
		var va string
		err := database.DBPNBP.
			Table("invoice_relations").
			Select("virtual_accounts.virtual_account").
			Joins("INNER JOIN virtual_accounts ON virtual_accounts.id = invoice_relations.virtual_account_id").
			Where("invoice_relations.detail_cicilan_id = ?", *payment.DetailCicilanID).
			Order("invoice_relations.id DESC").
			Limit(1).
			Pluck("virtual_accounts.virtual_account", &va).Error

		if err == nil && va != "" {
			virtualAccount = va
		}
	} else if payment.Source == "registrasi" && payment.RegistrasiID != nil {
		// Query langsung dari invoice_relations ke virtual_accounts
		var va string
		err := database.DBPNBP.
			Table("invoice_relations").
			Select("virtual_accounts.virtual_account").
			Joins("INNER JOIN virtual_accounts ON virtual_accounts.id = invoice_relations.virtual_account_id").
			Where("invoice_relations.registrasi_mahasiswa_id = ?", *payment.RegistrasiID).
			Order("invoice_relations.id DESC").
			Limit(1).
			Pluck("virtual_accounts.virtual_account", &va).Error

		if err == nil && va != "" {
			virtualAccount = va
		}
	}

	return virtualAccount
}

// GenerateInvoicePDF GET /api/v1/invoice-pdf
// Generate PDF invoice untuk pembayaran
// Query params:
//   - payment_id: ID dari payment (detail_cicilan_id atau registrasi_id)
//   - source: "cicilan" atau "registrasi"
func GenerateInvoicePDF(c *gin.Context) {
	mhswMaster, mustreturn := getMahasiswa(c)
	if mustreturn {
		return
	}

	if mhswMaster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data mahasiswa tidak ditemukan"})
		return
	}

	paymentIDStr := c.Query("payment_id")
	source := c.Query("source")

	if paymentIDStr == "" || source == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Parameter tidak valid",
			"message": "Harus menyertakan payment_id dan source",
		})
		return
	}

	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Parameter tidak valid",
			"message": "payment_id harus berupa angka",
		})
		return
	}

	tagihanRepo := repositories.NewTagihanRepository(database.DBPNBP, database.DBPNBP)
	activeYear, err := tagihanRepo.GetActiveFinanceYear()
	if err != nil {
		utils.Log.Error("Gagal mengambil finance year aktif", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tahun aktif tidak ditemukan"})
		return
	}

	dummyMahasiswa := &models.Mahasiswa{
		MhswID: mhswMaster.StudentID,
		Nama:   mhswMaster.NamaLengkap,
	}

	tagihanNewService := services.NewTagihanNewService(*tagihanRepo)
	historyList, err := tagihanNewService.GetHistoryTagihanMahasiswa(dummyMahasiswa, activeYear)
	if err != nil {
		utils.Log.Error("Gagal mengambil riwayat pembayaran", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pembayaran"})
		return
	}

	// Cari payment yang sesuai
	var payment *models.TagihanResponse
	for i := range historyList {
		if historyList[i].ID == uint(paymentID) && historyList[i].Source == source {
			payment = &historyList[i]
			break
		}
	}

	if payment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data pembayaran tidak ditemukan"})
		return
	}

	// Ambil virtual account
	virtualAccount := getVirtualAccount(payment)

	// Ambil nama prodi
	prodiName := ""
	if mhswMaster.ProdiID > 0 {
		var prodi models.ProdiPnbp
		if err := database.DBPNBP.Where("id = ?", mhswMaster.ProdiID).First(&prodi).Error; err == nil {
			prodiName = prodi.NamaProdi
		}
	}

	// Generate PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header dengan logo - gunakan embedded logo atau file system
	var logoData []byte
	var logoName string

	// Prioritas 1: Gunakan embedded logo jika ada
	if len(embeddedLogo) > 0 {
		logoData = embeddedLogo
		logoName = "embedded_logo.png"
	} else {
		// Prioritas 2: Cari logo dari file system
		logoPath := findLogoPath()
		if logoPath != "" {
			if fileData, err := os.ReadFile(logoPath); err == nil {
				logoData = fileData
				logoName = logoPath
			}
		}
	}

	// Header dengan logo di kiri dan teks di kanan
	startY := pdf.GetY()
	logoHeight := 0.0
	logoWidth := 0.0

	// Tambahkan logo ke PDF jika ada (di kiri, lebih kecil)
	if len(logoData) > 0 {
		// Register image dari byte data menggunakan RegisterImageOptionsReader
		opt := gofpdf.ImageOptions{
			ImageType: "PNG",
		}

		// Register image dari bytes reader
		reader := bytes.NewReader(logoData)
		infoPtr := pdf.RegisterImageOptionsReader(logoName, opt, reader)
		if infoPtr != nil {
			// Calculate size (max width 25mm untuk logo di header, maintain aspect ratio)
			maxWidth := 15.0
			imgWidth := infoPtr.Width()
			imgHeight := infoPtr.Height()

			width := imgWidth
			height := imgHeight

			if imgWidth > maxWidth {
				ratio := maxWidth / imgWidth
				width = maxWidth
				height = imgHeight * ratio
			}

			logoWidth = width
			logoHeight = height

			// Logo di kiri, dengan margin top sedikit
			x := 12.0
			pdf.ImageOptions(logoName, x, startY, width, height, false, opt, 0, "")
		}
	}

	// Teks institusi di samping logo (mulai dari x = 45mm jika ada logo, atau 10mm jika tidak ada)
	textStartX := 30.0
	if logoWidth == 0 {
		textStartX = 10.0
	}

	pdf.SetXY(textStartX, startY)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 6, "UNIVERSITAS SILIWANGI", "", 0, "L", false, 0, "")
	pdf.Ln(7)

	pdf.SetX(textStartX)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(0, 4, "Jl. Siliwangi No. 24, Tasikmalaya, Jawa Barat 46115", "", 0, "L", false, 0, "")
	pdf.Ln(5)

	pdf.SetX(textStartX)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(0, 4, "Telp: (0265) 330634 | Email: info@unsil.ac.id", "", 0, "L", false, 0, "")

	// Set Y position ke bawah logo atau teks (ambil yang lebih besar)
	if logoHeight > 0 {
		nextY := startY + logoHeight + 5
		if pdf.GetY() < nextY {
			pdf.SetY(nextY)
		}
	} else {
		pdf.Ln(3)
	}

	// Garis pemisah
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(4)

	// Judul "BUKTI PEMBAYARAN" di tengah
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(0, 8, "BUKTI PEMBAYARAN", "", 0, "C", false, 0, "")
	pdf.Ln(10)

	// Line separator
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(10)

	// Informasi Mahasiswa
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(95, 8, "Informasi Mahasiswa")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(95, 8, "Informasi Pembayaran")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)

	// Baris 1: Nama | Virtual Account
	pdf.Cell(95, 6, fmt.Sprintf("Nama: %s", mhswMaster.NamaLengkap))
	if virtualAccount != "" {
		pdf.Cell(95, 6, fmt.Sprintf("Virtual Account: %s", virtualAccount))
	} else {
		pdf.Cell(95, 6, "") // Spacer jika tidak ada virtual account
	}
	pdf.Ln(6)

	// Baris 2: NPM/NIM | Tanggal Bayar
	studentId := payment.NPM
	if studentId == "" {
		studentId = mhswMaster.StudentID
	}
	pdf.Cell(95, 6, fmt.Sprintf("NPM/NIM: %s", studentId))
	pdf.Cell(95, 6, fmt.Sprintf("Tanggal Bayar: %s", formatDateForPDF(payment.UpdatedAt)))
	pdf.Ln(6)

	// Baris 3: Prodi | Status
	if prodiName != "" {
		pdf.Cell(95, 6, fmt.Sprintf("Prodi: %s", prodiName))
	} else {
		pdf.Cell(95, 6, "Prodi: -")
	}

	statusText := "LUNAS"
	if payment.Status == "partial" {
		statusText = "SEBAGIAN"
	}
	pdf.Cell(95, 6, fmt.Sprintf("Status: %s", statusText))
	pdf.Ln(6)

	// Baris 4: Tahun Akademik | Angsuran/Jenis Tagihan
	pdf.Cell(95, 6, fmt.Sprintf("Tahun Akademik: %s", payment.AcademicYear))

	// Tampilkan angsuran untuk cicilan, atau jenis tagihan untuk registrasi
	if payment.Source == "cicilan" && payment.SequenceNo != nil {
		pdf.Cell(95, 6, fmt.Sprintf("Angsuran: Ke-%d", *payment.SequenceNo))
	} else if payment.Source == "registrasi" {
		pdf.Cell(95, 6, "Tagihan: Registrasi")
	} else {
		pdf.Cell(95, 6, "") // Spacer jika tidak ada
	}
	pdf.Ln(10)

	// Detail Tagihan
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Detail Tagihan")
	pdf.Ln(8)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(245, 245, 245)
	pdf.CellFormat(140, 8, "Keterangan", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 8, "Jumlah", "1", 0, "R", true, 0, "")
	pdf.Ln(8)

	// Table rows
	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(140, 8, payment.BillName, "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 8, formatCurrencyForPDF(payment.Amount), "1", 0, "R", true, 0, "")
	pdf.Ln(8)

	if payment.Beasiswa > 0 {
		pdf.CellFormat(140, 8, "Beasiswa", "1", 0, "L", true, 0, "")
		pdf.CellFormat(50, 8, "- "+formatCurrencyForPDF(payment.Beasiswa), "1", 0, "R", true, 0, "")
		pdf.Ln(8)
	}

	if payment.BantuanUKT > 0 {
		pdf.CellFormat(140, 8, "Bantuan UKT", "1", 0, "L", true, 0, "")
		pdf.CellFormat(50, 8, "- "+formatCurrencyForPDF(payment.BantuanUKT), "1", 0, "R", true, 0, "")
		pdf.Ln(8)
	}

	pdf.Ln(10)

	// Total Section
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(140, 6, "", "", 0, "", false, 0, "")
	pdf.CellFormat(50, 6, fmt.Sprintf("Total Tagihan: %s", formatCurrencyForPDF(payment.Amount)), "", 0, "R", false, 0, "")
	pdf.Ln(6)

	if payment.Beasiswa > 0 {
		pdf.CellFormat(140, 6, "", "", 0, "", false, 0, "")
		pdf.CellFormat(50, 6, fmt.Sprintf("Beasiswa: -%s", formatCurrencyForPDF(payment.Beasiswa)), "", 0, "R", false, 0, "")
		pdf.Ln(6)
	}

	if payment.BantuanUKT > 0 {
		pdf.CellFormat(140, 6, "", "", 0, "", false, 0, "")
		pdf.CellFormat(50, 6, fmt.Sprintf("Bantuan UKT: -%s", formatCurrencyForPDF(payment.BantuanUKT)), "", 0, "R", false, 0, "")
		pdf.Ln(6)
	}

	pdf.CellFormat(140, 6, "", "", 0, "", false, 0, "")
	pdf.CellFormat(50, 6, fmt.Sprintf("Total Dibayar: %s", formatCurrencyForPDF(payment.PaidAmount)), "", 0, "R", false, 0, "")
	pdf.Ln(6)

	if payment.RemainingAmount > 0 {
		pdf.CellFormat(140, 6, "", "", 0, "", false, 0, "")
		pdf.CellFormat(50, 6, fmt.Sprintf("Sisa Tagihan: %s", formatCurrencyForPDF(payment.RemainingAmount)), "", 0, "R", false, 0, "")
		pdf.Ln(6)
	}

	pdf.Ln(15)

	// Terms & Conditions / Catatan Penting
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 6, "CATATAN PENTING", "", 0, "L", false, 0, "")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(60, 60, 60)

	// Terms & Conditions text
	terms := []string{
		"1. Dokumen ini adalah bukti pembayaran yang sah dan dapat digunakan sebagai referensi pembayaran.",
		"2. Simpan dokumen ini dengan baik untuk keperluan administrasi dan audit.",
		"3. Jika terdapat perbedaan data, harap segera menghubungi bagian keuangan universitas.",
		"4. Dokumen ini dicetak secara otomatis oleh sistem dan tidak memerlukan tanda tangan manual.",
		"5. Untuk informasi lebih lanjut, hubungi bagian keuangan di Telp: (0265) 330634 atau Email: info@unsil.ac.id",
	}

	for _, term := range terms {
		pdf.CellFormat(0, 4, term, "", 0, "L", false, 0, "")
		pdf.Ln(4)
	}

	pdf.Ln(5)

	// Footer - Tanggal cetak
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(102, 102, 102)
	pdf.CellFormat(0, 5, fmt.Sprintf("Dicetak pada: %s", formatDateForPDF(time.Now())), "", 0, "C", false, 0, "")
	pdf.Ln(3)
	pdf.CellFormat(0, 5, "EPNBP - Universitas Siliwangi", "", 0, "C", false, 0, "")

	// Set response headers
	c.Header("Content-Type", "application/pdf")
	filename := fmt.Sprintf("Invoice_%s_%s_%d.pdf",
		payment.BillName,
		studentId,
		time.Now().Unix(),
	)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Output PDF
	err = pdf.Output(c.Writer)
	if err != nil {
		utils.Log.Error("Gagal generate PDF", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate PDF"})
		return
	}
}

// formatCurrencyForPDF format currency untuk PDF
func formatCurrencyForPDF(amount int64) string {
	return fmt.Sprintf("Rp %s", formatNumber(amount))
}

// formatNumber format number dengan separator ribuan
func formatNumber(n int64) string {
	str := fmt.Sprintf("%d", n)
	result := ""
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += "."
		}
		result += string(char)
	}
	return result
}

// formatDateForPDF format date untuk PDF
func formatDateForPDF(t time.Time) string {
	months := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	return fmt.Sprintf("%d %s %d, %02d:%02d",
		t.Day(),
		months[t.Month()-1],
		t.Year(),
		t.Hour(),
		t.Minute(),
	)
}

// findLogoPath mencari path logo.png
// Mencari di beberapa lokasi yang mungkin:
// 1. Environment variable LOGO_PATH
// 2. Backend directory (./logo.png) - PRIORITAS
// 3. Root project (../logo.png dari backend)
// 4. Current directory (./logo.png)
func findLogoPath() string {
	// Cek environment variable
	if logoPath := os.Getenv("LOGO_PATH"); logoPath != "" {
		if _, err := os.Stat(logoPath); err == nil {
			return logoPath
		}
	}

	// Cek di backend directory (prioritas utama)
	backendPath := "logo.png"
	if _, err := os.Stat(backendPath); err == nil {
		absPath, err := filepath.Abs(backendPath)
		if err == nil {
			return absPath
		}
	}

	// Cek di root project (satu level di atas backend)
	rootPath := filepath.Join("..", "logo.png")
	if _, err := os.Stat(rootPath); err == nil {
		absPath, err := filepath.Abs(rootPath)
		if err == nil {
			return absPath
		}
	}

	// Cek di current directory
	currentPath := filepath.Join(".", "logo.png")
	if _, err := os.Stat(currentPath); err == nil {
		absPath, err := filepath.Abs(currentPath)
		if err == nil {
			return absPath
		}
	}

	return ""
}
