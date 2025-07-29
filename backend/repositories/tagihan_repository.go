package repositories

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
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
	var bills []models.StudentBill
	err := r.DB.
		Where("student_id = ? AND academic_year <> ? and ( (quantity * amount ) - paid_amount) > 0 ", studentID, academicYear).
		Order("created_at ASC").
		Find(&bills).Error
	return bills, err
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
