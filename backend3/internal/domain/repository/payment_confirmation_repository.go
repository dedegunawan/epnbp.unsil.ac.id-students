package repository

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/internal/domain/entity"
)

// PaymentConfirmationRepository defines the interface for payment confirmation operations
type PaymentConfirmationRepository interface {
	// Create creates a new payment confirmation
	Create(confirmation *entity.PaymentConfirmation) error
	
	// FindByStudentBillID finds payment confirmations by student bill ID
	FindByStudentBillID(studentBillID uint) ([]entity.PaymentConfirmation, error)
	
	// FindByID finds a payment confirmation by ID
	FindByID(id uint) (*entity.PaymentConfirmation, error)
}




