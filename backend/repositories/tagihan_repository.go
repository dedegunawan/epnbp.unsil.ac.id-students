package repositories

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"gorm.io/gorm"
	"strconv"
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

// Ambil FinanceYear aktif
func (r *TagihanRepository) GetActiveFinanceYear() (*models.FinanceYear, error) {
	var fy models.FinanceYear
	if err := r.DB.Where("is_active = ?", true).First(&fy).Error; err != nil {
		return nil, err
	}
	// override finance year dari budget_period
	return &fy, nil
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

	utils.Log.Info("Override Finance Year untuk mahasiswa:", mahasiswa.MhswID)

	prodi, fakultas, errProdi, errFakultas := r.GetValidProdiPnbp(mahasiswa.Prodi.KodeProdi)

	utils.Log.Info("Prodi:", prodi.KodeProdi, " - ", prodi.NamaProdi)
	utils.Log.Info("Fakultas:", fakultas.KodeFakultas, " - ", fakultas.NamaFakultas)
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
	if !skippedFakultas {
		utils.Log.Info("Lakukan override fakultas:", fakultas.KodeFakultas, " - ", fakultas.NamaFakultas)
		overrideFakultas, err := r.GetBudgetOverride("fakultas", strconv.Itoa(int(fakultas.ID)), strconv.Itoa(int(budgetPeriod.ID)))
		if err == nil && overrideFakultas != nil {
			defaultPaymentStartDate = overrideFakultas.PaymentStartDate
			defaultPaymentEndDate = overrideFakultas.PaymentEndDate
		}
	}

	// cek apakah ada prodi mahasiswa di daftar override
	if !skippedProdi {
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

	err2 := r.DBPNBP.
		Where("kode_fakultas = ?", ProdiID[:2]).First(&fakultas).Error

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
