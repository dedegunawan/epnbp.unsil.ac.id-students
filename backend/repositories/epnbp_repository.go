package repositories

import (
	"time"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"gorm.io/gorm"
)

type EpnbpRepository interface {
	GetDB() *gorm.DB
	FindNotExpiredByStudentBill(studentBillID string) (*models.PayUrl, error)
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
