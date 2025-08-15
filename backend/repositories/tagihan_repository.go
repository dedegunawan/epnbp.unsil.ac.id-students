package repositories

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"gorm.io/gorm"
)

type TagihanRepository struct {
	DB *gorm.DB
}

func NewTagihanRepository(db *gorm.DB) *TagihanRepository {
	return &TagihanRepository{DB: db}
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
	return &fy, nil
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
		Select("academic_year, SUM((quantity * amount) - paid_amount) AS total_unpaid").
		Where("student_id = ? AND academic_year <> ? AND ((quantity * amount) - paid_amount) > 0", studentID, academicYear).
		Group("academic_year").
		Order("academic_year ASC").
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

	// Ambil ulang semua tagihan detail
	var bills []models.StudentBill
	err = r.DB.
		Where("student_id = ? AND academic_year <> ? AND ((quantity * amount) - paid_amount) > 0", studentID, academicYear).
		Order("created_at ASC").
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
	err := r.DB.
		Where("student_id = ? AND academic_year <> ? and ( (quantity * amount ) - paid_amount) <= 0 ", studentID, academicYear).
		Order("created_at ASC").
		Find(&bills).Error
	return bills, err
}
