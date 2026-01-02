package repositories

import (
	"fmt"
	"time"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"gorm.io/gorm"
)

type EpnbpRepository interface {
	GetDB() *gorm.DB
	FindNotExpiredByStudentBill(studentBillID string) (*models.PayUrl, error)
	FindByInvoiceId(invoiceID string) (*models.PayUrl, error)
	GetUnpaidPayUrls(limit int) ([]models.PayUrl, error)
	CheckInvoiceStatusInMySQL(invoiceID uint) (bool, *time.Time, error)
}

type epnbpRepository struct {
	DB *gorm.DB
}

func NewEpnbpRepository(db *gorm.DB) EpnbpRepository {
	return &epnbpRepository{DB: db}
}

func (s *epnbpRepository) GetDB() *gorm.DB {
	return s.DB
}

func (er *epnbpRepository) FindNotExpiredByStudentBill(studentBillID string) (*models.PayUrl, error) {
	var payUrl models.PayUrl

	err := er.DB.
		Where("student_bill_id = ?", studentBillID).
		Where("expired_at IS NULL OR expired_at > ?", time.Now()).
		Order("expired_at ASC").
		First(&payUrl).Error

	if err != nil {
		return nil, err
	}

	return &payUrl, nil
}

func (er *epnbpRepository) FindByInvoiceId(invoiceID string) (*models.PayUrl, error) {
	var payUrl models.PayUrl

	err := er.DB.
		Where("invoice_id = ?", invoiceID).
		First(&payUrl).Error

	if err != nil {
		return nil, err
	}

	return &payUrl, nil
}

// GetUnpaidPayUrls mengambil pay_urls yang belum terbayar
// dengan join ke student_bills untuk cek apakah masih ada sisa tagihan
func (er *epnbpRepository) GetUnpaidPayUrls(limit int) ([]models.PayUrl, error) {
	var payUrls []models.PayUrl

	err := er.DB.
		Joins("JOIN student_bills ON student_bills.id = pay_urls.student_bill_id").
		Where("(student_bills.quantity * student_bills.amount) - student_bills.paid_amount > 0").
		Where("pay_urls.expired_at IS NULL OR pay_urls.expired_at > ?", time.Now()).
		Order("pay_urls.created_at ASC").
		Limit(limit).
		Find(&payUrls).Error

	return payUrls, err
}

// CheckInvoiceStatusInMySQL mengecek status invoice di database MySQL (DBPNBP)
// Method ini tidak digunakan langsung, karena memerlukan DBPNBP
// Implementasi ada di payment_status_worker.go
func (er *epnbpRepository) CheckInvoiceStatusInMySQL(invoiceID uint) (bool, *time.Time, error) {
	// Method ini didefinisikan untuk interface compatibility
	// Implementasi sebenarnya ada di payment_status_worker.go
	return false, nil, fmt.Errorf("method harus dipanggil dengan DBPNBP dari service layer")
}
