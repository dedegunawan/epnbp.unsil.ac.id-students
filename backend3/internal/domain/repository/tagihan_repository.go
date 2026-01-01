package repository

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/internal/domain/entity"
)

// TagihanRepository defines the interface for tagihan (student bill) operations
type TagihanRepository interface {
	// GetStudentBills gets all student bills for a student in a specific academic year
	GetStudentBills(mhswID string, academicYear string) ([]entity.StudentBill, error)
	
	// GetAllUnpaidBillsExcept gets all unpaid bills except for the specified academic year
	GetAllUnpaidBillsExcept(mhswID string, academicYear string) ([]entity.StudentBill, error)
	
	// GetAllPaidBillsExcept gets all paid bills except for the specified academic year
	GetAllPaidBillsExcept(mhswID string, academicYear string) ([]entity.StudentBill, error)
	
	// FindStudentBillByID finds a student bill by ID
	FindStudentBillByID(studentBillID string) (*entity.StudentBill, error)
	
	// DeleteUnpaidBills deletes unpaid bills for a student in a specific academic year
	DeleteUnpaidBills(mhswID string, academicYear string) error
	
	// GetActiveFinanceYearWithOverride gets the active finance year with override for a student
	GetActiveFinanceYearWithOverride(mahasiswa entity.Mahasiswa) (*entity.BudgetPeriod, error)
	
	// Create creates a new student bill
	Create(studentBill *entity.StudentBill) error
	
	// Update updates an existing student bill
	Update(studentBill *entity.StudentBill) error
}


