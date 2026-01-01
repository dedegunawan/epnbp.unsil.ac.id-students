package repository

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/internal/domain/entity"
)

// EpnbpRepository defines the interface for payment URL operations
type EpnbpRepository interface {
	// FindNotExpiredByStudentBill finds a payment URL that is not expired for a student bill
	FindNotExpiredByStudentBill(studentBillID string) (*entity.PayUrl, error)
	
	// Create creates a new payment URL
	Create(payUrl *entity.PayUrl) error
	
	// Update updates an existing payment URL
	Update(payUrl *entity.PayUrl) error
	
	// FindByInvoiceID finds payment URL by invoice ID
	FindByInvoiceID(invoiceID uint) (*entity.PayUrl, error)
}


