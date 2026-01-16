package repositories

import (
	"strconv"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"gorm.io/gorm"
)

type TagihanRepository struct {
	DB     *gorm.DB
	DBPNBP *gorm.DB // Tambahkan DBPNBP jika diperlukan
}

func NewTagihanRepository(db *gorm.DB, DPPNBP *gorm.DB) *TagihanRepository {
	return &TagihanRepository{DB: db, DBPNBP: DPPNBP}
}

func (r *TagihanRepository) FindStudentBillByID(studentBillID string) (*models.StudentBill, error) {
	var studentBill models.StudentBill
	err := r.DB.First(&studentBill, "ID = ?", studentBillID).Error
	return &studentBill, err
}

// Ambil FinanceYear aktif dari budget_periods (bukan dari finance_years)
func (r *TagihanRepository) GetActiveFinanceYear() (*models.FinanceYear, error) {
	var budgetPeriod models.BudgetPeriod
	// Ambil dari budget_periods yang aktif
	if err := r.DBPNBP.Where("is_active = ?", true).First(&budgetPeriod).Error; err != nil {
		return nil, err
	}
	
	// Convert BudgetPeriod ke FinanceYear
	fy := &models.FinanceYear{
		Code:            budgetPeriod.Kode,
		Description:     budgetPeriod.Name,
		AcademicYear:    budgetPeriod.Kode,
		FiscalYear:      budgetPeriod.FiscalYear,
		FiscalSemester:  strconv.Itoa(budgetPeriod.Semester),
		StartDate:       budgetPeriod.PaymentStartDate,
		EndDate:         budgetPeriod.PaymentEndDate,
		IsActive:        budgetPeriod.IsActive,
	}
	
	return fy, nil
}

func (r *TagihanRepository) GetActiveFinanceYearWithOverride(mahasiswa models.Mahasiswa) (*models.FinanceYear, error) {
	financeYear, err := r.GetActiveFinanceYear()
	if err == nil {
		err = r.OverrideFinanceYear(financeYear, mahasiswa)
	}
	return financeYear, err
}

func (r *TagihanRepository) OverrideFinanceYear(financeYear *models.FinanceYear, mahasiswa models.Mahasiswa) error {
	// ambil dari daftar budget_period
	budgetPeriod, err := r.GetActiveBudgetPeriod()
	utils.Log.Info("Mencoba melakukan override")

	if err != nil {
		utils.Log.Info("Tidak melakukan override finance year karena tidak ada budget period aktif:", err)
		return nil
	}

	if budgetPeriod == nil {
		utils.Log.Info("Tidak melakukan override finance year karena tidak ada budget period aktif, budget periode null")
		return nil
	}

	skippedFakultas := false
	skippedProdi := false

	// Validasi MhswID tidak kosong
	if mahasiswa.MhswID == "" {
		utils.Log.Warn("OverrideFinanceYear: MhswID kosong, skip override")
		return nil
	}

	utils.Log.Info("Override Finance Year untuk mahasiswa:", mahasiswa.MhswID)

	// Ambil prodi langsung dari mahasiswa_masters di database PNBP
	var mhswMaster models.MahasiswaMaster
	errMhswMaster := r.DBPNBP.Where("student_id = ?", mahasiswa.MhswID).First(&mhswMaster).Error

	var prodi *models.ProdiPnbp
	var fakultas *models.FakultasPnbp
	var errProdi, errFakultas error

	if errMhswMaster == nil && mhswMaster.ProdiID > 0 {
		// Ambil prodi dari database PNBP berdasarkan ProdiID dari mahasiswa_masters
		var prodiData models.ProdiPnbp
		errProdi = r.DBPNBP.Where("id = ?", mhswMaster.ProdiID).First(&prodiData).Error
		if errProdi == nil {
			prodi = &prodiData
			// Ambil fakultas dari database PNBP berdasarkan FakultasID dari prodi
			var fakultasData models.FakultasPnbp
			errFakultas = r.DBPNBP.Where("id = ?", prodi.FakultasID).First(&fakultasData).Error
			if errFakultas == nil {
				fakultas = &fakultasData
			}
			utils.Log.Info("Prodi diambil dari mahasiswa_masters di database PNBP", map[string]interface{}{
				"mhswID":       mahasiswa.MhswID,
				"ProdiID":      mhswMaster.ProdiID,
				"KodeProdi":    prodi.KodeProdi,
				"FakultasID":   prodi.FakultasID,
				"KodeFakultas": fakultas.KodeFakultas,
			})
		} else {
			utils.Log.Warn("Gagal ambil prodi dari mahasiswa_masters, fallback ke kode prodi", map[string]interface{}{
				"mhswID":  mahasiswa.MhswID,
				"ProdiID": mhswMaster.ProdiID,
				"error":   errProdi,
			})
			// Fallback: coba ambil dari kode_prodi jika ada di FullData
			if prodiIDString := utils.GetStringFromAny(mahasiswa.ParseFullData()["ProdiID"]); prodiIDString != "" {
				prodi, fakultas, errProdi, errFakultas = r.GetValidProdiPnbp(prodiIDString)
			}
		}
	} else {
		// Fallback: ambil dari kode_prodi di FullData atau Prodi lokal
		prodiIDString := ""
		if prodiIDString = utils.GetStringFromAny(mahasiswa.ParseFullData()["ProdiID"]); prodiIDString == "" && mahasiswa.Prodi.KodeProdi != "" {
			prodiIDString = mahasiswa.Prodi.KodeProdi
		}
		if prodiIDString != "" {
			utils.Log.Info("Fallback: ambil prodi dari kode_prodi", "mhswID", mahasiswa.MhswID, "kodeProdi", prodiIDString)
			prodi, fakultas, errProdi, errFakultas = r.GetValidProdiPnbp(prodiIDString)
		} else {
			utils.Log.Warn("Tidak dapat mengambil prodi dari mahasiswa_masters maupun kode_prodi", "mhswID", mahasiswa.MhswID)
			errProdi = gorm.ErrRecordNotFound
			errFakultas = gorm.ErrRecordNotFound
		}
	}

	if prodi != nil {
		utils.Log.Info("Prodi:", prodi.KodeProdi, " - ", prodi.NamaProdi)
	} else {
		utils.Log.Warn("Prodi nil")
	}
	if fakultas != nil {
		utils.Log.Info("Fakultas:", fakultas.KodeFakultas, " - ", fakultas.NamaFakultas)
	} else {
		utils.Log.Warn("Fakultas nil")
	}
	utils.Log.Info("Err Prodi:", errProdi)
	utils.Log.Info("Err Fakultas:", errFakultas)

	if errProdi != nil {
		skippedProdi = true
	}
	if errFakultas != nil {
		skippedFakultas = true
	}

	defaultPaymentStartDate := budgetPeriod.PaymentStartDate
	defaultPaymentEndDate := budgetPeriod.PaymentEndDate

	// cek apakah ada fakultas mahasiswa di daftar override
	if !skippedFakultas && fakultas != nil {
		utils.Log.Info("Lakukan override fakultas:", fakultas.KodeFakultas, " - ", fakultas.NamaFakultas)
		overrideFakultas, err := r.GetBudgetOverride("fakultas", strconv.Itoa(int(fakultas.ID)), strconv.Itoa(int(budgetPeriod.ID)))
		if err == nil && overrideFakultas != nil {
			defaultPaymentStartDate = overrideFakultas.PaymentStartDate
			defaultPaymentEndDate = overrideFakultas.PaymentEndDate
		}
	}

	// cek apakah ada prodi mahasiswa di daftar override
	if !skippedProdi && prodi != nil {
		utils.Log.Info("Lakukan override prodi:", prodi.KodeProdi, " - ", prodi.NamaProdi)
		overrideProdi, err := r.GetBudgetOverride("prodi", strconv.Itoa(int(prodi.ID)), strconv.Itoa(int(budgetPeriod.ID)))
		if err == nil && overrideProdi != nil {
			defaultPaymentStartDate = overrideProdi.PaymentStartDate
			defaultPaymentEndDate = overrideProdi.PaymentEndDate
		}
	}

	// cek apakah mahasiswa ada di daftar override
	findMahasiswaPnbp, err := r.GetScopeIDByMhswID(mahasiswa.MhswID)
	var overrideIndividual *models.BudgetPeriodPaymentOverride
	if err == nil && findMahasiswaPnbp != nil {
		overrideIndividual, err = r.GetBudgetOverride("individual", strconv.Itoa(int(findMahasiswaPnbp.ID)), strconv.Itoa(int(budgetPeriod.ID)))
	}
	utils.Log.Info("Err override individual:", err, overrideIndividual)
	if err == nil && overrideIndividual != nil {
		utils.Log.Info("Lakukan override individual:", mahasiswa.MhswID)
		defaultPaymentStartDate = overrideIndividual.PaymentStartDate
		defaultPaymentEndDate = overrideIndividual.PaymentEndDate
	}
	financeYear.Code = budgetPeriod.Kode
	financeYear.Description = budgetPeriod.Name
	financeYear.AcademicYear = budgetPeriod.Kode
	financeYear.FiscalYear = budgetPeriod.FiscalYear
	financeYear.FiscalSemester = strconv.Itoa(budgetPeriod.Semester)
	financeYear.StartDate = defaultPaymentStartDate
	financeYear.EndDate = defaultPaymentEndDate
	financeYear.IsActive = budgetPeriod.IsActive

	return nil
}

func (r *TagihanRepository) GetActiveBudgetPeriod() (*models.BudgetPeriod, error) {
	var fy models.BudgetPeriod
	if err := r.DBPNBP.Where("is_active = ?", true).First(&fy).Error; err != nil {
		return nil, err
	}
	// override finance year dari budget_period
	return &fy, nil
}
func (r *TagihanRepository) GetBudgetOverride(scope string, scopeID string, budget_period_id string) (*models.BudgetPeriodPaymentOverride, error) {
	var fy models.BudgetPeriodPaymentOverride
	if err := r.DBPNBP.Where("scope_type = ?", scope).
		Where("budget_period_id = ?", budget_period_id).
		Where("scope_id = ?", scopeID).
		Where("is_active = ?", true).
		First(&fy).Error; err != nil {
		return nil, err
	}
	// override finance year dari budget_period
	return &fy, nil
}

func (r *TagihanRepository) GetValidProdiPnbp(ProdiID string) (*models.ProdiPnbp, *models.FakultasPnbp, error, error) {
	utils.Log.Info("ProdiID:", ProdiID)
	var prodi models.ProdiPnbp
	var fakultas models.FakultasPnbp
	err := r.DBPNBP.
		Where("kode_prodi = ?", ProdiID).First(&prodi).Error

	var err2 error
	// Validasi panjang ProdiID sebelum melakukan slice
	if len(ProdiID) >= 2 {
		kodeFakultas := ProdiID[:2]
		err2 = r.DBPNBP.
			Where("kode_fakultas = ?", kodeFakultas).First(&fakultas).Error
	} else {
		utils.Log.Warn("ProdiID terlalu pendek untuk mengambil kode fakultas", "ProdiID", ProdiID, "length", len(ProdiID))
		err2 = gorm.ErrRecordNotFound
	}

	return &prodi, &fakultas, err, err2

}

func (r *TagihanRepository) GetScopeIDByMhswID(MhswID string) (*models.MahasiswaPnbp, error) {
	var mahasiswa models.MahasiswaPnbp
	err := r.DBPNBP.
		Where("MhswID = ?", MhswID).First(&mahasiswa).Error
	return &mahasiswa, err
}

// Ambil tagihan mahasiswa berdasarkan student_id & academic_year
func (r *TagihanRepository) GetStudentBills(studentID string, academicYear string) ([]models.StudentBill, error) {
	var bills []models.StudentBill
	err := r.DB.
		Where("student_id = ? AND academic_year = ?", studentID, academicYear).
		Order("created_at ASC").
		Find(&bills).Error
	return bills, err
}

// Ambil tagihan mahasiswa berdasarkan student_id & academic_year
func (r *TagihanRepository) GetAllUnpaidBillsExcept(studentID string, academicYear string) ([]models.StudentBill, error) {
	// Ambil total unpaid per tahun akademik (kecuali tahun sekarang)
	type TahunUnpaid struct {
		AcademicYear string
		TotalUnpaid  int64
	}

	var tahunUnpaidList []TahunUnpaid
	err := r.DB.
		Table("student_bills").
		Select("student_bills.academic_year, SUM((student_bills.quantity * student_bills.amount) - student_bills.paid_amount) AS total_unpaid").
		Joins("INNER JOIN finance_years ON finance_years.academic_year = student_bills.academic_year").
		Where("student_bills.student_id = ? AND student_bills.academic_year <> ? AND ((student_bills.quantity * student_bills.amount) - student_bills.paid_amount) > 0", studentID, academicYear).
		Where("finance_years.is_active = ?", true).
		Group("student_bills.academic_year").
		Order("student_bills.academic_year ASC").
		Scan(&tahunUnpaidList).Error
	if err != nil {
		return nil, err
	}

	// Ambil total beasiswa dari tahun akademik yang dikecualikan
	beasiswa := r.GetBeasiswaByMahasiswaTahun(studentID, academicYear)

	// Simulasi pengurangan beasiswa terhadap tagihan tiap tahun
	unpaidAfterBeasiswa := make(map[string]int64)

	for _, item := range tahunUnpaidList {
		if beasiswa >= item.TotalUnpaid {
			unpaidAfterBeasiswa[item.AcademicYear] = 0
			beasiswa -= item.TotalUnpaid
		} else {
			unpaidAfterBeasiswa[item.AcademicYear] = item.TotalUnpaid - beasiswa
			beasiswa = 0
		}
	}

	// Ambil ulang semua tagihan detail - hanya dari finance year yang aktif
	var bills []models.StudentBill
	err = r.DB.Model(&models.StudentBill{}).
		Joins("INNER JOIN finance_years ON finance_years.academic_year = student_bills.academic_year").
		Where("student_bills.student_id = ? AND student_bills.academic_year <> ? AND ((student_bills.quantity * student_bills.amount) - student_bills.paid_amount) > 0", studentID, academicYear).
		Where("finance_years.is_active = ?", true).
		Order("student_bills.created_at ASC").
		Find(&bills).Error
	if err != nil {
		return bills, err
	}

	// Filter tagihan hanya yang masih punya unpaid setelah dikurangi beasiswa
	var filteredBills []models.StudentBill
	for _, bill := range bills {
		sisa := (int64(bill.Quantity) * bill.Amount) - bill.PaidAmount
		if unpaidAfterBeasiswa[bill.AcademicYear] > 0 {
			if unpaidAfterBeasiswa[bill.AcademicYear] >= sisa {
				filteredBills = append(filteredBills, bill)
				unpaidAfterBeasiswa[bill.AcademicYear] -= sisa
			} else {
				bill.PaidAmount += sisa - unpaidAfterBeasiswa[bill.AcademicYear]
				filteredBills = append(filteredBills, bill)
				unpaidAfterBeasiswa[bill.AcademicYear] = 0
			}
		}
	}

	return filteredBills, nil
}

func (r *TagihanRepository) GetBeasiswaByMahasiswaTahun(studentID string, academicYear string) int64 {
	var total int64

	dbEpnbp := database.DBPNBP

	err := dbEpnbp.Table("detail_beasiswa").
		Joins("JOIN beasiswa ON beasiswa.id = detail_beasiswa.beasiswa_id").
		Select("CAST(COALESCE(SUM(detail_beasiswa.nominal_beasiswa), 0) AS UNSIGNED)").
		Where("beasiswa.status = ?", "active").
		Where("detail_beasiswa.tahun_id = ?", academicYear).
		Where("detail_beasiswa.npm = ?", studentID).
		Scan(&total).Error

	if err != nil {
		utils.Log.Info("Error saat ambil total nominal_beasiswa:", err)
		return 0
	}

	return total

}

// Ambil tagihan mahasiswa berdasarkan student_id & academic_year
func (r *TagihanRepository) DeleteUnpaidBills(studentID string, academicYear string) error {
	err := r.DB.
		Where("student_id = ? AND academic_year = ? and paid_amount = 0 ", studentID, academicYear).
		Delete(&models.StudentBill{}).Error
	return err
}

// Ambil tagihan mahasiswa berdasarkan student_id & academic_year
func (r *TagihanRepository) GetAllPayUrlByStudentBillID(studentBillID uint) ([]models.PayUrl, error) {
	var payUrls []models.PayUrl
	err := r.DB.
		Where(`student_bill_id = ?`, studentBillID).
		Order("created_at ASC").
		Find(&payUrls).Error
	return payUrls, err
}

// Ambil tagihan mahasiswa berdasarkan student_id & academic_year
func (r *TagihanRepository) GetAllPaidBillsExcept(studentID string, academicYear string) ([]models.StudentBill, error) {
	var bills []models.StudentBill
	err := r.DB.Model(&models.StudentBill{}).
		Joins("INNER JOIN finance_years ON finance_years.academic_year = student_bills.academic_year").
		Where("student_bills.student_id = ? AND student_bills.academic_year <> ? AND ((student_bills.quantity * student_bills.amount) - student_bills.paid_amount) <= 0", studentID, academicYear).
		Where("finance_years.is_active = ?", true).
		Order("student_bills.created_at ASC").
		Find(&bills).Error
	return bills, err
}

// Ambil tagihan mahasiswa berdasarkan student_id & academic_year
func (r *TagihanRepository) GetTotalStudentBill(studentID string, academicYear string) int64 {
	var total int64
	err := r.DB.
		Model(&models.StudentBill{}).
		Where("student_id = ? AND academic_year = ?", studentID, academicYear).
		Select("SUM((quantity * amount))").
		Scan(&total).Error

	if err != nil {
		total = 0
	}
	return total
}
