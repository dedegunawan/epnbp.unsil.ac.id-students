package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
)

type TagihanService interface {
	CreateNewTagihan(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error

	CreateNewTagihanPasca(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error
	HitungSemesterSaatIni(tahunIDAwal string, tahunIDSekarang string) (int, error)
	SavePaymentConfirmation(studentBill models.StudentBill, vaNumber string, paymentDate string, objectName string) (*models.PaymentConfirmation, error)

	CekCicilanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool
	CekPenangguhanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool
	CekBeasiswaMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool
	CekDepositMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool
	IsNominalDibayarLebihKecilSeharusnya(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) (bool, int64, int64)
	CreateNewTagihanSekurangnya(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear, tagihanKurang int64) error
	ValidateBillAmount(studentBill *models.StudentBill, mahasiswa *models.Mahasiswa) error
}

type tagihanService struct {
	repo                    repositories.TagihanRepository
	masterTagihanRepository repositories.MasterTagihanRepository
}

func NewTagihanService(repo repositories.TagihanRepository, masterTagihanRepository repositories.MasterTagihanRepository) TagihanService {
	return &tagihanService{repo: repo, masterTagihanRepository: masterTagihanRepository}
}

func (r *tagihanService) GetNominalBeasiswa(studentId string, academicYear string) int64 {
	var total int64

	dbEpnbp := database.DBPNBP

	err := dbEpnbp.Table("detail_beasiswa").
		Joins("JOIN beasiswa ON beasiswa.id = detail_beasiswa.beasiswa_id").
		Select("COALESCE(CAST(SUM(detail_beasiswa.nominal_beasiswa) AS SIGNED), 0)").
		Where("beasiswa.status = ?", "active").
		Where("detail_beasiswa.tahun_id = ?", academicYear).
		Where("detail_beasiswa.npm = ?", studentId).
		Scan(&total).Error

	if err != nil {
		utils.Log.Info("Error saat ambil total nominal_beasiswa:", err)
		return 0
	}

	return total

}

func (r *tagihanService) CheckDepositMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) {
	//dbPnbp := database.DBPNBP

	// cek apakah sudah ada deposit yang digunakan di tahun tersebut
	//var deposit models.DepositLedgerEntry
	//dbPnbp.Where("student_id = ? AND academic_year = ? AND status = ?",)

	// jika sudah ada kembalikan hasilnya & sukses, kecuali masih ada kekurangan, buatkan tagihan baru nya

	// jika belum ada & masih punya deposit, buatkan tagihan deposit baru untuk mahasiswa tersebut

	// jika tidak punya deposit kembalikan hasil kosong & lanjutkan
}

func (r *tagihanService) GenerateCicilanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {
	mhswID := string(mahasiswa.MhswID)
	financeCode := financeYear.Code
	dbEpnbp := database.DBPNBP

	var cicilanJatuhTempo []models.DetailCicilan
	today := time.Now().Format("2006-01-02") // Format YYYY-MM-DD

	err := dbEpnbp.Preload("Cicilan").
		Joins("JOIN cicilans ON cicilans.id = detail_cicilans.cicilan_id").
		Where("detail_cicilans.due_date <= ?", today).
		Where("cicilans.tahun_id = ? AND cicilans.npm = ?", financeCode, mhswID).
		Find(&cicilanJatuhTempo).Error

	if err == nil && len(cicilanJatuhTempo) > 0 {
		for _, data := range cicilanJatuhTempo {
			dt := models.StudentBill{
				StudentID:          string(mahasiswa.MhswID),
				AcademicYear:       financeYear.AcademicYear,
				BillTemplateItemID: 0,
				Name:               "Cicilan UKT",
				Amount:             data.Amount,
				PaidAmount:         0,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}
			r.repo.DB.Create(&dt)
		}
		return true
	}
	return false
}

func (r *tagihanService) HasCicilanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {
	mhswID := string(mahasiswa.MhswID)
	financeCode := financeYear.Code
	dbEpnbp := database.DBPNBP

	var hasCicilanCount int64

	err := dbEpnbp.Preload("Cicilan").
		Joins("JOIN cicilans ON cicilans.id = detail_cicilans.cicilan_id").
		Where("cicilans.tahun_id = ? AND cicilans.npm = ?", financeCode, mhswID).
		Count(&hasCicilanCount).Error

	if err == nil && hasCicilanCount > 0 {
		return true
	}
	return false
}

// getUKTFromMahasiswaMasters mengambil kelompok UKT (kel_ukt) dari mahasiswa_masters di database PNBP
// Catatan: mhswMaster.UKT adalah nominal (int64), tapi yang dibutuhkan adalah kelompok UKT (string: "1"-"7")
func (r *tagihanService) getUKTFromMahasiswaMasters(mhswID string) (string, error) {
	var mhswMaster models.MahasiswaMaster
	err := database.DBPNBP.Where("student_id = ?", mhswID).First(&mhswMaster).Error
	if err != nil {
		return "", err
	}

	// Ambil kelompok UKT dari detail_tagihan berdasarkan MasterTagihanID dan nominal UKT
	if mhswMaster.MasterTagihanID != 0 {
		var detailTagihan models.DetailTagihan
		errDetail := database.DBPNBP.Where("master_tagihan_id = ? AND nominal = ?", mhswMaster.MasterTagihanID, mhswMaster.UKT).
			First(&detailTagihan).Error
		if errDetail == nil && detailTagihan.KelUKT != nil {
			return *detailTagihan.KelUKT, nil
		}

		// Fallback: cari berdasarkan MasterTagihanID saja
		errFallback := database.DBPNBP.Where("master_tagihan_id = ?", mhswMaster.MasterTagihanID).
			First(&detailTagihan).Error
		if errFallback == nil && detailTagihan.KelUKT != nil {
			return *detailTagihan.KelUKT, nil
		}
	}

	// Fallback terakhir: gunakan nominal sebagai string (untuk kompatibilitas)
	return strconv.Itoa(int(mhswMaster.UKT)), nil
}

// getBIPOTIDFromMahasiswaMasters mengambil BIPOTID langsung dari mahasiswa_masters -> master_tagihan di database PNBP
func (r *tagihanService) getBIPOTIDFromMahasiswaMasters(mhswID string) (string, error) {
	var mhswMaster models.MahasiswaMaster
	err := database.DBPNBP.Preload("MasterTagihan").Where("student_id = ?", mhswID).First(&mhswMaster).Error
	if err != nil {
		utils.Log.Warn("Mahasiswa tidak ditemukan di mahasiswa_masters", "mhswID", mhswID, "error", err)
		return "", fmt.Errorf("mahasiswa tidak ditemukan di mahasiswa_masters: %w", err)
	}

	if mhswMaster.MasterTagihanID == 0 {
		utils.Log.Warn("MasterTagihanID = 0 untuk mahasiswa", "mhswID", mhswID)
		return "", fmt.Errorf("master_tagihan_id tidak ditemukan untuk mahasiswa %s (MasterTagihanID=0)", mhswID)
	}

	if mhswMaster.MasterTagihan == nil {
		utils.Log.Warn("MasterTagihan nil untuk mahasiswa, mencoba load manual", "mhswID", mhswID, "MasterTagihanID", mhswMaster.MasterTagihanID)
		// Coba load manual jika Preload gagal
		var masterTagihan models.MasterTagihan
		errLoad := database.DBPNBP.Where("id = ?", mhswMaster.MasterTagihanID).First(&masterTagihan).Error
		if errLoad != nil {
			return "", fmt.Errorf("gagal load master_tagihan untuk mahasiswa %s: %w", mhswID, errLoad)
		}
		mhswMaster.MasterTagihan = &masterTagihan
	}

	if mhswMaster.MasterTagihan.BipotID == 0 {
		utils.Log.Warn("BipotID = 0 di master_tagihan", "mhswID", mhswID, "MasterTagihanID", mhswMaster.MasterTagihanID)
		return "", fmt.Errorf("bipotid tidak ditemukan di master_tagihan untuk mahasiswa %s (BipotID=0)", mhswID)
	}

	BIPOTID := strconv.Itoa(int(mhswMaster.MasterTagihan.BipotID))
	utils.Log.Info("BIPOTID diambil dari mahasiswa_masters", "mhswID", mhswID, "BIPOTID", BIPOTID, "MasterTagihanID", mhswMaster.MasterTagihanID)
	return BIPOTID, nil
}

func (r *tagihanService) CreateNewTagihan(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error {
	utils.Log.Info("CreateNewTagihan dimulai", map[string]interface{}{
		"mhswID":       mahasiswa.MhswID,
		"nama":         mahasiswa.Nama,
		"BIPOTID":      mahasiswa.BIPOTID,
		"UKT":          mahasiswa.UKT,
		"academicYear": financeYear.AcademicYear,
	})

	// interception: jika mahasiswa memiliki data cicilan generate dari cicilan tersebut
	hasCicilan := r.GenerateCicilanMahasiswa(mahasiswa, financeYear)
	if hasCicilan {
		utils.Log.Info("Mahasiswa memiliki cicilan, generate dari cicilan", "mhswID", mahasiswa.MhswID)
		return nil
	}

	// Ambil data dari mahasiswa_masters -> master_tagihan -> detail_tagihan
	// JANGAN gunakan mahasiswa.BIPOTID atau mahasiswa.UKT, ambil langsung dari mahasiswa_masters
	utils.Log.Info("Mengambil data dari mahasiswa_masters", "mhswID", mahasiswa.MhswID)

	var mhswMaster models.MahasiswaMaster
	err := database.DBPNBP.Preload("MasterTagihan").Where("student_id = ?", mahasiswa.MhswID).First(&mhswMaster).Error
	if err != nil {
		utils.Log.Error("Mahasiswa tidak ditemukan di mahasiswa_masters", "mhswID", mahasiswa.MhswID, "error", err)
		return fmt.Errorf("mahasiswa tidak ditemukan di mahasiswa_masters untuk %s: %w", mahasiswa.MhswID, err)
	}

	if mhswMaster.MasterTagihanID == 0 {
		utils.Log.Error("MasterTagihanID = 0 untuk mahasiswa", "mhswID", mahasiswa.MhswID)
		return fmt.Errorf("master_tagihan_id tidak ditemukan untuk mahasiswa %s (MasterTagihanID=0)", mahasiswa.MhswID)
	}

	// Load master_tagihan jika belum ter-load
	if mhswMaster.MasterTagihan == nil {
		utils.Log.Info("MasterTagihan nil, mencoba load manual", "mhswID", mahasiswa.MhswID, "MasterTagihanID", mhswMaster.MasterTagihanID)
		var masterTagihan models.MasterTagihan
		errLoad := database.DBPNBP.Where("id = ?", mhswMaster.MasterTagihanID).First(&masterTagihan).Error
		if errLoad != nil {
			utils.Log.Error("Gagal load master_tagihan", "mhswID", mahasiswa.MhswID, "MasterTagihanID", mhswMaster.MasterTagihanID, "error", errLoad)
			return fmt.Errorf("gagal load master_tagihan untuk mahasiswa %s: %w", mahasiswa.MhswID, errLoad)
		}
		mhswMaster.MasterTagihan = &masterTagihan
	}

	// Catatan: BIPOTID tidak diperlukan lagi karena kita tidak menggunakan bill_template
	// Langsung menggunakan detail_tagihan dari master_tagihan_id
	if mhswMaster.MasterTagihan.BipotID == 0 {
		utils.Log.Warn("BipotID = 0 di master_tagihan, lanjut tanpa BIPOTID (tidak diperlukan)", map[string]interface{}{
			"mhswID":          mahasiswa.MhswID,
			"MasterTagihanID": mhswMaster.MasterTagihanID,
		})
	} else {
		BIPOTID := strconv.Itoa(int(mhswMaster.MasterTagihan.BipotID))
		utils.Log.Info("BIPOTID diambil dari master_tagihan (untuk referensi)", map[string]interface{}{
			"mhswID":          mahasiswa.MhswID,
			"BIPOTID":         BIPOTID,
			"MasterTagihanID": mhswMaster.MasterTagihanID,
		})
	}

	// Ambil kelompok UKT dari mahasiswa_masters.ukt
	// mahasiswa_masters.ukt = detail_tagihan.kel_ukt (harus sama persis)
	// Gunakan CAST untuk membandingkan float dengan string di database
	var UKT string // Kelompok UKT (kel_ukt) dari mahasiswa_masters
	var errDetail error
	var detailTagihan models.DetailTagihan

	// Coba beberapa format untuk mencocokkan dengan kel_ukt di database
	// Format 1: int sebagai string ("2")
	kelompokUKTInt := strconv.Itoa(int(mhswMaster.UKT))
	kelompokUKTFloat := fmt.Sprintf("%.2f", mhswMaster.UKT)
	kelompokUKTNoDecimal := fmt.Sprintf("%.0f", mhswMaster.UKT)

	triedFormats := []string{kelompokUKTInt, kelompokUKTFloat, kelompokUKTNoDecimal}

	errDetail = database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, kelompokUKTInt).
		First(&detailTagihan).Error

	if errDetail != nil {
		// Format 2: float dengan 2 desimal ("2.00")
		errDetail = database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, kelompokUKTFloat).
			First(&detailTagihan).Error
		if errDetail == nil {
			UKT = kelompokUKTFloat
		}
	} else {
		UKT = kelompokUKTInt
	}

	if errDetail != nil {
		// Format 3: tanpa desimal ("2")
		errDetail = database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, kelompokUKTNoDecimal).
			First(&detailTagihan).Error
		if errDetail == nil {
			UKT = kelompokUKTNoDecimal
		}
	}

	if errDetail != nil {
		// JANGAN fallback ke semua master_tagihan_id karena akan mengambil yang pertama (bisa salah)
		// Gunakan nilai dari mahasiswa_masters sebagai string
		UKT = strconv.Itoa(int(mhswMaster.UKT))
		utils.Log.Warn("Kelompok UKT tidak ditemukan dengan format apapun, menggunakan nilai dari mahasiswa_masters", map[string]interface{}{
			"mhswID":          mahasiswa.MhswID,
			"UKT":             UKT,
			"uktValue":        mhswMaster.UKT,
			"masterTagihanID": mhswMaster.MasterTagihanID,
			"error":           errDetail.Error(),
			"triedFormats":    triedFormats,
		})
	} else {
		utils.Log.Info("Kelompok UKT ditemukan dari detail_tagihan", map[string]interface{}{
			"mhswID":               mahasiswa.MhswID,
			"kelompokUKT":          UKT,
			"uktValue":             mhswMaster.UKT,
			"masterTagihanID":      mhswMaster.MasterTagihanID,
			"detailTagihanID":      detailTagihan.ID,
			"detailTagihanNominal": detailTagihan.Nominal,
		})
	}

	// Ambil semua detail_tagihan dari master_tagihan_id yang sesuai dengan UKT nominal
	// JANGAN gunakan bill_template atau bill_template_items, langsung dari detail_tagihan
	utils.Log.Info("Mencari detail_tagihan dari master_tagihan", map[string]interface{}{
		"masterTagihanID": mhswMaster.MasterTagihanID,
		"uktValue":        mhswMaster.UKT, // Ini adalah kelompok UKT, bukan nominal
		"kelompokUKT":     UKT,
		"mhswID":          mahasiswa.MhswID,
	})

	var detailTagihans []models.DetailTagihan
	// Ambil semua detail_tagihan yang sesuai dengan master_tagihan_id dan kel_ukt
	// JANGAN gunakan DISTINCT dengan field list karena bisa menyebabkan field lain tidak ter-populate
	// Gunakan query biasa dan filter duplikasi di aplikasi jika perlu
	// Pastikan query menggunakan UKT yang sudah ditemukan (bukan dari mahasiswa_masters langsung)
	errDetailList := database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, UKT).
		Find(&detailTagihans).Error

	// Log query yang digunakan - PASTIKAN UKT sudah benar
	utils.Log.Info("Query detail_tagihan untuk generate bill", map[string]interface{}{
		"masterTagihanID": mhswMaster.MasterTagihanID,
		"kelompokUKT":     UKT,
		"uktValue":        mhswMaster.UKT,
		"mhswID":          mahasiswa.MhswID,
		"query":           fmt.Sprintf("master_tagihan_id = %d AND kel_ukt = '%s'", mhswMaster.MasterTagihanID, UKT),
		"note":            "UKT ini harus match dengan kel_ukt di database untuk mendapatkan nominal yang benar",
	})

	if errDetailList != nil {
		utils.Log.Error("Gagal query detail_tagihan", map[string]interface{}{
			"masterTagihanID": mhswMaster.MasterTagihanID,
			"kelompokUKT":     UKT,
			"error":           errDetailList.Error(),
		})
		return fmt.Errorf("gagal query detail_tagihan untuk master_tagihan_id %d dan kel_ukt %s: %w", mhswMaster.MasterTagihanID, UKT, errDetailList)
	}

	// Jika tidak ditemukan dengan kel_ukt, jangan fallback ke semua master_tagihan_id
	// karena itu akan menghasilkan semua tagihan, bukan hanya yang sesuai kelompok UKT
	if len(detailTagihans) == 0 {
		utils.Log.Error("Detail tagihan tidak ditemukan dengan kel_ukt yang sesuai", map[string]interface{}{
			"masterTagihanID": mhswMaster.MasterTagihanID,
			"kelompokUKT":     UKT,
			"uktValue":        mhswMaster.UKT,
			"mhswID":          mahasiswa.MhswID,
		})
		return fmt.Errorf("tidak ada detail_tagihan yang ditemukan untuk master_tagihan_id %d dan kel_ukt %s (mahasiswa %s)", mhswMaster.MasterTagihanID, UKT, mahasiswa.MhswID)
	}

	// Log jumlah detail_tagihan yang ditemukan untuk debugging
	// Log detail setiap record untuk memastikan Nominal ter-populate
	var detailTagihansLog []map[string]interface{}
	for i, dt := range detailTagihans {
		detailTagihansLog = append(detailTagihansLog, map[string]interface{}{
			"Index":           i,
			"ID":              dt.ID,
			"MasterTagihanID": dt.MasterTagihanID,
			"KelUKT":          dt.KelUKT,
			"Nama":            dt.Nama,
			"Nominal":         dt.Nominal,
		})
		// Log individual untuk memastikan Nominal ter-populate
		utils.Log.Info(fmt.Sprintf("Detail tagihan [%d] setelah query", i), map[string]interface{}{
			"ID":              dt.ID,
			"MasterTagihanID": dt.MasterTagihanID,
			"KelUKT":          dt.KelUKT,
			"Nama":            dt.Nama,
			"Nominal":         dt.Nominal,
		})
	}
	utils.Log.Info("Detail tagihan ditemukan", map[string]interface{}{
		"count":           len(detailTagihans),
		"masterTagihanID": mhswMaster.MasterTagihanID,
		"kelompokUKT":     UKT,
		"mhswID":          mahasiswa.MhswID,
		"detailTagihans":  detailTagihansLog,
	})

	// Jika ada lebih dari 1 record, ambil hanya yang pertama (atau filter berdasarkan nama tertentu)
	// Biasanya untuk UKT hanya ada 1 record per kelompok UKT
	if len(detailTagihans) > 1 {
		utils.Log.Warn("Ditemukan lebih dari 1 detail_tagihan, akan menggunakan yang pertama", map[string]interface{}{
			"count":           len(detailTagihans),
			"masterTagihanID": mhswMaster.MasterTagihanID,
			"kelompokUKT":     UKT,
		})
		// Filter untuk mengambil hanya record dengan nama "UKT" atau "Uang Kuliah"
		// Prioritas: 1. Nama mengandung "UKT", 2. Nama mengandung "UANG KULIAH", 3. Yang pertama
		var filteredDetailTagihans []models.DetailTagihan
		var uktOnlyDetailTagihans []models.DetailTagihan

		for _, dt := range detailTagihans {
			// Prioritas 1: Nama mengandung "UKT"
			if strings.Contains(strings.ToUpper(dt.Nama), "UKT") {
				uktOnlyDetailTagihans = append(uktOnlyDetailTagihans, dt)
			}
			// Prioritas 2: Nama mengandung "Uang Kuliah"
			if strings.Contains(strings.ToUpper(dt.Nama), "UANG KULIAH") {
				filteredDetailTagihans = append(filteredDetailTagihans, dt)
			}
		}

		// Gunakan yang mengandung "UKT" jika ada, jika tidak gunakan "UANG KULIAH", jika tidak gunakan yang pertama
		if len(uktOnlyDetailTagihans) > 0 {
			detailTagihans = uktOnlyDetailTagihans
			utils.Log.Info("Detail tagihan setelah filter (prioritas UKT)", map[string]interface{}{
				"count":    len(detailTagihans),
				"selected": detailTagihans[0].Nama,
				"nominal":  detailTagihans[0].Nominal,
			})
		} else if len(filteredDetailTagihans) > 0 {
			detailTagihans = filteredDetailTagihans
			utils.Log.Info("Detail tagihan setelah filter (prioritas UANG KULIAH)", map[string]interface{}{
				"count":    len(detailTagihans),
				"selected": detailTagihans[0].Nama,
				"nominal":  detailTagihans[0].Nominal,
			})
		} else {
			// Ambil yang pertama, tapi log warning
			utils.Log.Warn("Tidak ada detail_tagihan dengan nama UKT atau UANG KULIAH, menggunakan yang pertama", map[string]interface{}{
				"count":    len(detailTagihans),
				"selected": detailTagihans[0].Nama,
				"nominal":  detailTagihans[0].Nominal,
				"allNames": func() []string {
					names := make([]string, len(detailTagihans))
					for i, dt := range detailTagihans {
						names[i] = dt.Nama
					}
					return names
				}(),
			})
			detailTagihans = detailTagihans[:1] // Ambil hanya yang pertama
		}
	}

	nominalBeasiswa := r.GetNominalBeasiswa(string(mahasiswa.MhswID), financeYear.AcademicYear)
	utils.Log.Info("nominalBeasiswa:", nominalBeasiswa)

	sisaBeasiswa := nominalBeasiswa
	// Generate StudentBill langsung dari detail_tagihan
	// Cek dulu apakah sudah ada StudentBill dengan name yang sama untuk tahun akademik ini
	for _, dt := range detailTagihans {
		// Cek apakah sudah ada StudentBill dengan student_id, academic_year, dan name yang sama
		var existingBill models.StudentBill
		errCheck := r.repo.DB.Where("student_id = ? AND academic_year = ? AND name = ?",
			mahasiswa.MhswID, financeYear.AcademicYear, dt.Nama).
			First(&existingBill).Error

		if errCheck == nil {
			// Sudah ada, cek apakah nominalnya sesuai dengan detail_tagihan
			// Amount seharusnya = detail_tagihan.Nominal - beasiswa yang sudah ada
			nominalTagihanSeharusnya := dt.Nominal
			if existingBill.Beasiswa > 0 {
				if existingBill.Beasiswa >= dt.Nominal {
					nominalTagihanSeharusnya = 0
				} else {
					nominalTagihanSeharusnya = dt.Nominal - existingBill.Beasiswa
				}
			}

			// Cek apakah Amount berbeda dengan yang seharusnya
			if existingBill.Amount != nominalTagihanSeharusnya {
				utils.Log.Warn("StudentBill sudah ada tapi Amount tidak sesuai, akan di-update", map[string]interface{}{
					"mhswID":               mahasiswa.MhswID,
					"academicYear":         financeYear.AcademicYear,
					"name":                 dt.Nama,
					"billID":               existingBill.ID,
					"amountLama":           existingBill.Amount,
					"amountBaru":           nominalTagihanSeharusnya,
					"detailTagihanNominal": dt.Nominal,
					"beasiswa":             existingBill.Beasiswa,
					"paidAmount":           existingBill.PaidAmount,
				})

				// Update Amount dengan nilai yang benar dari detail_tagihan
				updateData := map[string]interface{}{
					"amount": nominalTagihanSeharusnya,
				}

				// Jika PaidAmount lebih besar dari Amount baru, set PaidAmount = Amount baru
				// Ini untuk mencegah PaidAmount > Amount
				if existingBill.PaidAmount > nominalTagihanSeharusnya {
					updateData["paid_amount"] = nominalTagihanSeharusnya
					utils.Log.Warn("PaidAmount lebih besar dari Amount baru, akan disesuaikan", map[string]interface{}{
						"mhswID":         mahasiswa.MhswID,
						"billID":         existingBill.ID,
						"paidAmountLama": existingBill.PaidAmount,
						"paidAmountBaru": nominalTagihanSeharusnya,
					})
				}

				errUpdate := r.repo.DB.Model(&existingBill).Updates(updateData).Error
				if errUpdate != nil {
					utils.Log.Error("Gagal update StudentBill", map[string]interface{}{
						"mhswID": mahasiswa.MhswID,
						"billID": existingBill.ID,
						"error":  errUpdate.Error(),
					})
				} else {
					utils.Log.Info("StudentBill berhasil di-update dengan nominal yang benar", map[string]interface{}{
						"mhswID":       mahasiswa.MhswID,
						"academicYear": financeYear.AcademicYear,
						"name":         dt.Nama,
						"billID":       existingBill.ID,
						"amountLama":   existingBill.Amount,
						"amountBaru":   nominalTagihanSeharusnya,
					})
				}
			} else {
				utils.Log.Info("StudentBill sudah ada dan nominal sudah sesuai, skip", map[string]interface{}{
					"mhswID":       mahasiswa.MhswID,
					"academicYear": financeYear.AcademicYear,
					"name":         dt.Nama,
					"billID":       existingBill.ID,
					"amount":       existingBill.Amount,
					"beasiswa":     existingBill.Beasiswa,
				})
			}
			continue
		}
		nominalBeasiswaSaatIni := int64(0)
		nominalTagihan := dt.Nominal

		// VALIDASI: Pastikan kelUKT match dengan UKT yang dicari
		// Jika tidak match, skip detail_tagihan ini (tidak digunakan untuk generate bill)
		kelUKTStr := ""
		if dt.KelUKT != nil {
			kelUKTStr = *dt.KelUKT
		}
		isMatch := kelUKTStr == UKT

		if !isMatch {
			utils.Log.Warn("SKIP: kelUKT tidak match dengan yang dicari, detail_tagihan ini tidak akan digunakan", map[string]interface{}{
				"mhswID":          mahasiswa.MhswID,
				"kelUKTSearched":  UKT,
				"kelUKTFound":     kelUKTStr,
				"nominal":         dt.Nominal,
				"detailTagihanID": dt.ID,
			})
			continue // Skip detail_tagihan yang tidak match
		}

		// Log detail tagihan yang akan digunakan untuk generate bill
		utils.Log.Info("Menggunakan detail_tagihan untuk generate StudentBill", map[string]interface{}{
			"mhswID":          mahasiswa.MhswID,
			"detailTagihanID": dt.ID,
			"nama":            dt.Nama,
			"kelUKT":          kelUKTStr,
			"kelUKTSearched":  UKT,
			"isMatch":         true,
			"nominal":         dt.Nominal,
			"masterTagihanID": dt.MasterTagihanID,
			"academicYear":    financeYear.AcademicYear,
		})

		if sisaBeasiswa > 0 && sisaBeasiswa >= dt.Nominal {
			sisaBeasiswa = sisaBeasiswa - dt.Nominal
			nominalBeasiswaSaatIni = dt.Nominal
			nominalTagihan = 0
		} else if sisaBeasiswa > 0 {
			nominalBeasiswaSaatIni = sisaBeasiswa
			nominalTagihan = dt.Nominal - nominalBeasiswaSaatIni
		}

		bill := models.StudentBill{
			StudentID:          string(mahasiswa.MhswID),
			AcademicYear:       financeYear.AcademicYear,
			BillTemplateItemID: 0, // Tidak menggunakan bill_template_item lagi
			Name:               dt.Nama,
			Amount:             nominalTagihan,
			Beasiswa:           nominalBeasiswaSaatIni,
			PaidAmount:         0,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := r.repo.DB.Create(&bill).Error; err != nil {
			utils.Log.Error("Gagal membuat StudentBill", map[string]interface{}{
				"mhswID":          mahasiswa.MhswID,
				"detailTagihanID": dt.ID,
				"nama":            dt.Nama,
				"nominal":         dt.Nominal,
				"error":           err.Error(),
			})
			return fmt.Errorf("gagal membuat tagihan mahasiswa dari detail_tagihan ID %d: %w", dt.ID, err)
		}

		utils.Log.Info("StudentBill berhasil dibuat dari detail_tagihan", map[string]interface{}{
			"mhswID":          mahasiswa.MhswID,
			"detailTagihanID": dt.ID,
			"nama":            dt.Nama,
			"nominal":         dt.Nominal,
			"amount":          nominalTagihan,
			"beasiswa":        nominalBeasiswaSaatIni,
		})
	}

	return nil
}
func (r *tagihanService) CreateNewTagihanPasca(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error {
	mhswID := mahasiswa.MhswID
	activeYear := financeYear.AcademicYear // Menggunakan AcademicYear sebagai tahun_id

	utils.Log.Info("CreateNewTagihanPasca: Mengambil data tagihan dari registrasi_mahasiswa", map[string]interface{}{
		"mhswID":     mhswID,
		"activeYear": activeYear,
	})

	// Ambil data tagihan dari tabel registrasi_mahasiswa di database PNBP
	// Filter: tahun_id = activeYear dan npm = MhswID
	var registrasiTagihan []models.RegistrasiMahasiswa
	if err := database.DBPNBP.
		Where("tahun_id = ?", activeYear).
		Where("npm = ?", mhswID).
		Find(&registrasiTagihan).Error; err != nil {
		return fmt.Errorf("gagal mengambil data tagihan dari registrasi_mahasiswa: %w", err)
	}

	if len(registrasiTagihan) == 0 {
		utils.Log.Info("CreateNewTagihanPasca: Tidak ada data tagihan ditemukan", map[string]interface{}{
			"mhswID":     mhswID,
			"activeYear": activeYear,
		})
		return fmt.Errorf("tidak ada data tagihan ditemukan untuk npm %s dengan tahun_id %s", mhswID, activeYear)
	}

	utils.Log.Info("CreateNewTagihanPasca: Data tagihan ditemukan", map[string]interface{}{
		"mhswID":        mhswID,
		"activeYear":    activeYear,
		"jumlahTagihan": len(registrasiTagihan),
	})

	// Generate StudentBill berdasarkan data dari registrasi_mahasiswa
	for _, reg := range registrasiTagihan {
		// Skip jika sudah bayar atau nominal_ukt tidak ada
		if reg.SudahBayar {
			utils.Log.Info("CreateNewTagihanPasca: Skip tagihan yang sudah dibayar", map[string]interface{}{
				"mhswID":     mhswID,
				"activeYear": activeYear,
				"id":         reg.ID,
			})
			continue
		}

		if reg.NominalUKT == nil || *reg.NominalUKT == 0 {
			utils.Log.Warn("CreateNewTagihanPasca: Skip tagihan dengan nominal_ukt kosong atau 0", map[string]interface{}{
				"mhswID":     mhswID,
				"activeYear": activeYear,
				"id":         reg.ID,
			})
			continue
		}

		// Tentukan nama tagihan berdasarkan kel_ukt atau default
		namaTagihan := "Tagihan Registrasi"
		if reg.KelUKT != nil && *reg.KelUKT != "" {
			namaTagihan = fmt.Sprintf("UKT Kelompok %s", *reg.KelUKT)
		}

		// Konversi nominal_ukt dari float64 ke int64 (rupiah)
		nominal := int64(*reg.NominalUKT)

		// Set paid_amount jika sudah ada pembayaran
		paidAmount := int64(0)
		if reg.NominalBayar != nil {
			paidAmount = int64(*reg.NominalBayar)
		}

		// Buat StudentBill dari data registrasi_mahasiswa
		bill := models.StudentBill{
			StudentID:          mhswID,
			AcademicYear:       activeYear,
			BillTemplateItemID: 0, // Tidak menggunakan BillTemplateItem untuk pascasarjana dari registrasi_mahasiswa
			Name:               namaTagihan,
			Amount:             nominal,
			PaidAmount:         paidAmount,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := r.repo.DB.Create(&bill).Error; err != nil {
			utils.Log.Error("CreateNewTagihanPasca: Gagal membuat tagihan", map[string]interface{}{
				"mhswID":     mhswID,
				"activeYear": activeYear,
				"error":      err,
			})
			return fmt.Errorf("gagal membuat tagihan mahasiswa: %w", err)
		}

		utils.Log.Info("CreateNewTagihanPasca: Tagihan berhasil dibuat", map[string]interface{}{
			"mhswID":      mhswID,
			"activeYear":  activeYear,
			"namaTagihan": namaTagihan,
			"nominal":     nominal,
			"paidAmount":  paidAmount,
		})
	}

	return nil
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

// HitungSemesterSaatIni menghitung semester saat ini berdasarkan tahun masuk (angkatan) dan budget_periods.kode
// Rumus: semester = (tahun sekarang - tahun masuk) * 2 + semester sekarang - semester masuk + 1
// Contoh: budget_periods.kode = 20252, angkatan = 2024, semester masuk = 1 → semester = (2025 - 2024) * 2 + 2 - 1 + 1 = 4
// Contoh: budget_periods.kode = 20252, angkatan = 2025, semester masuk = 1 → semester = (2025 - 2025) * 2 + 2 - 1 + 1 = 2
func (r *tagihanService) HitungSemesterSaatIni(tahunIDAwal string, tahunIDSekarang string) (int, error) {
	utils.Log.Info("HitungSemesterSaatIni", map[string]interface{}{
		"tahunIDAwal":     tahunIDAwal,
		"tahunIDSekarang": tahunIDSekarang,
	})

	if len(tahunIDSekarang) != 5 {
		return 0, fmt.Errorf("format tahunIDSekarang tidak valid, harus 5 digit seperti 20251")
	}

	// Parsing tahun dan semester dari budget_periods.kode (tahunIDSekarang)
	tahunSekarang, err1 := strconv.Atoi(tahunIDSekarang[:4])
	semesterSekarang, err2 := strconv.Atoi(tahunIDSekarang[4:])

	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("gagal parsing tahun atau semester dari budget_periods.kode: %v, %v", err1, err2)
	}

	// Ambil tahun masuk (angkatan) dan semester masuk dari tahunIDAwal
	var tahunMasuk, semesterMasuk int
	if len(tahunIDAwal) == 5 {
		// Jika tahunIDAwal format YYYYS, ambil tahun dan semester masuknya
		tahunMasuk, err1 = strconv.Atoi(tahunIDAwal[:4])
		semesterMasuk, err2 = strconv.Atoi(tahunIDAwal[4:])
		if err1 != nil || err2 != nil {
			return 0, fmt.Errorf("gagal parsing tahun atau semester masuk dari tahunIDAwal: %v, %v", err1, err2)
		}
	} else {
		// Jika tahunIDAwal hanya tahun (4 digit), default semester masuk = 1
		tahunMasuk, err1 = strconv.Atoi(tahunIDAwal)
		semesterMasuk = 1 // Default semester masuk = 1 (Ganjil)
		if err1 != nil {
			return 0, fmt.Errorf("gagal parsing tahun masuk: %v", err1)
		}
	}

	// Rumus: semester = (tahun sekarang - tahun masuk) * 2 + semester sekarang - semester masuk + 1
	selisihTahun := tahunSekarang - tahunMasuk
	semester := (selisihTahun * 2) + semesterSekarang - semesterMasuk + 1

	utils.Log.Info("Perhitungan semester", map[string]interface{}{
		"tahunMasuk":       tahunMasuk,
		"semesterMasuk":    semesterMasuk,
		"tahunSekarang":    tahunSekarang,
		"semesterSekarang": semesterSekarang,
		"selisihTahun":     selisihTahun,
		"semester":         semester,
	})

	return semester, nil
}

func (r *tagihanService) SavePaymentConfirmation(studentBill models.StudentBill, vaNumber string, paymentDate string, objectName string) (*models.PaymentConfirmation, error) {
	paymentConfirmation := models.PaymentConfirmation{
		StudentBillID: studentBill.ID,
		VaNumber:      vaNumber,
		PaymentDate:   paymentDate,
		ObjectName:    objectName,
		Message:       "",
	}
	r.repo.DB.Save(&paymentConfirmation)

	// check all payment id is success or not
	payUrls, err := r.repo.GetAllPayUrlByStudentBillID(studentBill.ID)
	if err != nil {
		return nil, err
	}

	epnbpRepo := repositories.NewEpnbpRepository(r.repo.DB)
	eService := NewEpnbpService(epnbpRepo)

	var realPaymentDate *time.Time
	isPaid := false
	invoiceIds := []string{}
	for _, payUrl := range payUrls {
		invoiceId := strconv.FormatUint(uint64(payUrl.InvoiceID), 10)
		isPaid, realPaymentDate = eService.CheckStatusPaidByInvoiceID(invoiceId)
		invoiceIds = append(invoiceIds, invoiceId)
		if isPaid {
			break
		}
	}
	if !isPaid {
		isPaid, realPaymentDate = eService.CheckStatusPaidByVirtualAccount(vaNumber, invoiceIds)
	}

	if isPaid {
		r.savePaidStudentBill(studentBill, studentBill.Amount, *realPaymentDate, vaNumber, objectName)
		return &paymentConfirmation, nil
	}

	return nil, nil
}

func (r *tagihanService) savePaidStudentBill(studentBill models.StudentBill, amount int64, realPaymentDate time.Time, vaNumber string, objectName string) bool {
	studentBill.PaidAmount = amount
	r.repo.DB.Save(&studentBill)

	studentPayment := models.StudentPayment{
		StudentID:    string(studentBill.StudentID),
		AcademicYear: studentBill.AcademicYear,
		PaymentRef:   vaNumber,
		Amount:       amount,
		Bank:         "",
		Method:       "VA",
		Note:         objectName,
		Date:         realPaymentDate,
	}
	r.repo.DB.Save(&studentPayment)

	studentPaymentAllocation := models.StudentPaymentAllocation{
		StudentPaymentID: studentPayment.ID,
		StudentBillID:    studentBill.ID,
		Amount:           amount,
	}
	r.repo.DB.Save(&studentPaymentAllocation)

	return true

}

func (r *tagihanService) CekCicilanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {
	mhswID := string(mahasiswa.MhswID)
	financeCode := financeYear.AcademicYear
	var hasCicilanCount int64
	dbEpnbp := database.DBPNBP
	_ = dbEpnbp.Where("npm = ? AND tahun_id = ?", mhswID, financeCode).Model(&models.Cicilan{}).Count(&hasCicilanCount).Error

	if hasCicilanCount > 0 {
		return true
	}

	return false
}

func (r *tagihanService) CekPenangguhanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {

	mhswID := string(mahasiswa.MhswID)
	financeCode := financeYear.AcademicYear
	var hasDepositDebitCount int64
	dbEpnbp := database.DBPNBP
	err := dbEpnbp.Where("npm = ? AND tahun_id = ? and direction = ?", mhswID, financeCode, "debit").
		Model(&models.DepositLedgerEntry{}).Count(&hasDepositDebitCount).Error

	if err != nil {
		utils.Log.Error("Error checking deposit debit count:", err)
		return false

	}

	utils.Log.Info("Has Deposit Debit Count for Mahasiswa:", mhswID, "Finance Year:", financeCode, "Count:", hasDepositDebitCount)

	if hasDepositDebitCount > 0 {
		return true
	}

	return false
}

func (r *tagihanService) CekBeasiswaMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {

	mhswID := string(mahasiswa.MhswID)
	financeCode := financeYear.AcademicYear
	var hasBeasiswaCount int64
	dbEpnbp := database.DBPNBP
	_ = dbEpnbp.Where("npm = ? AND tahun_id = ?", mhswID, financeCode).
		Model(&models.DetailBeasiswa{}).Count(&hasBeasiswaCount).Error

	if hasBeasiswaCount > 0 {
		return true
	}

	return false
}

func (r *tagihanService) CekDepositMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) bool {
	return false
}

func (r *tagihanService) IsNominalDibayarLebihKecilSeharusnya(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) (bool, int64, int64) {
	// seharusnya diambil dari BillTemplateItem
	tagihanSeharusnya := r.masterTagihanRepository.GetNominalTagihanMahasiswa(*mahasiswa)

	// ambil nominal tagihan yang sudah dibayar oleh mahasiswa
	totalTagihanDibayar := r.repo.GetTotalStudentBill(mahasiswa.MhswID, financeYear.AcademicYear)
	utils.Log.Info("Tagihan seharusnya:", tagihanSeharusnya, " Total tagihan dibayar:", totalTagihanDibayar)

	return totalTagihanDibayar < tagihanSeharusnya, tagihanSeharusnya, totalTagihanDibayar
}

func (r *tagihanService) CreateNewTagihanSekurangnya(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear, tagihanKurang int64) error {
	studentBill := models.StudentBill{
		StudentID:          string(mahasiswa.MhswID),
		AcademicYear:       financeYear.AcademicYear,
		BillTemplateItemID: 0, // Asumsikan tidak ada item template yang
		Name:               "UKT",
		Amount:             tagihanKurang,
		PaidAmount:         0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	utils.Log.Info("Membuat tagihan mahasiswa dengan nominal kurang:", tagihanKurang)

	if err := r.repo.DB.Create(&studentBill).Error; err != nil {
		return fmt.Errorf("gagal membuat tagihan mahasiswa: %w", err)
	}

	return nil
}

// ValidateBillAmount memvalidasi apakah nominal tagihan sesuai dengan detail_tagihan
// Mengembalikan error jika nominal tidak sesuai
func (r *tagihanService) ValidateBillAmount(studentBill *models.StudentBill, mahasiswa *models.Mahasiswa) error {
	// Ambil detail_tagihan berdasarkan nama tagihan dan mahasiswa
	var mhswMaster models.MahasiswaMaster
	errMhswMaster := database.DBPNBP.Where("student_id = ?", mahasiswa.MhswID).First(&mhswMaster).Error
	if errMhswMaster != nil {
		utils.Log.Warn("ValidateBillAmount: mahasiswa_masters tidak ditemukan", map[string]interface{}{
			"mhswID": mahasiswa.MhswID,
			"error":  errMhswMaster.Error(),
		})
		// Jika tidak ditemukan, skip validasi (untuk kompatibilitas)
		return nil
	}

	if mhswMaster.MasterTagihanID == 0 {
		utils.Log.Warn("ValidateBillAmount: MasterTagihanID adalah 0", map[string]interface{}{
			"mhswID": mahasiswa.MhswID,
		})
		return nil
	}

	// Ambil UKT dari mahasiswa
	UKT := utils.GetStringFromAny(mahasiswa.UKT)
	if UKT == "" || UKT == "0" {
		utils.Log.Warn("ValidateBillAmount: UKT kosong atau 0", map[string]interface{}{
			"mhswID": mahasiswa.MhswID,
			"UKT":    UKT,
		})
		return nil
	}

	// Cari detail_tagihan berdasarkan master_tagihan_id, kel_ukt, dan nama tagihan
	var detailTagihan models.DetailTagihan
	errDetail := database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ? AND nama = ?",
		mhswMaster.MasterTagihanID, UKT, studentBill.Name).
		First(&detailTagihan).Error

	if errDetail != nil {
		// Jika tidak ditemukan dengan nama yang sama, coba cari berdasarkan master_tagihan_id dan kel_ukt saja
		var detailTagihans []models.DetailTagihan
		errDetailList := database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?",
			mhswMaster.MasterTagihanID, UKT).
			Find(&detailTagihans).Error

		if errDetailList != nil || len(detailTagihans) == 0 {
			utils.Log.Warn("ValidateBillAmount: detail_tagihan tidak ditemukan", map[string]interface{}{
				"mhswID":          mahasiswa.MhswID,
				"masterTagihanID": mhswMaster.MasterTagihanID,
				"UKT":             UKT,
				"billName":        studentBill.Name,
				"error":           errDetail.Error(),
			})
			// Jika tidak ditemukan, skip validasi (untuk kompatibilitas)
			return nil
		}

		// Gunakan yang pertama jika ada beberapa
		detailTagihan = detailTagihans[0]
	}

	// Hitung nominal tagihan yang seharusnya
	nominalTagihanSeharusnya := detailTagihan.Nominal
	if studentBill.Beasiswa > 0 {
		if studentBill.Beasiswa >= detailTagihan.Nominal {
			nominalTagihanSeharusnya = 0
		} else {
			nominalTagihanSeharusnya = detailTagihan.Nominal - studentBill.Beasiswa
		}
	}

	// Validasi: cek apakah Amount sesuai dengan yang seharusnya
	if studentBill.Amount != nominalTagihanSeharusnya {
		utils.Log.Warn("ValidateBillAmount: Amount tidak sesuai dengan detail_tagihan", map[string]interface{}{
			"mhswID":               mahasiswa.MhswID,
			"billID":               studentBill.ID,
			"billName":             studentBill.Name,
			"amountSaatIni":        studentBill.Amount,
			"amountSeharusnya":     nominalTagihanSeharusnya,
			"detailTagihanNominal": detailTagihan.Nominal,
			"beasiswa":             studentBill.Beasiswa,
		})
		return fmt.Errorf("nominal tagihan tidak sesuai. Silakan klik 'Perbaiki Tagihan' untuk memperbarui tagihan")
	}

	utils.Log.Info("ValidateBillAmount: Amount sesuai dengan detail_tagihan", map[string]interface{}{
		"mhswID":   mahasiswa.MhswID,
		"billID":   studentBill.ID,
		"billName": studentBill.Name,
		"amount":   studentBill.Amount,
		"beasiswa": studentBill.Beasiswa,
	})

	return nil
}
