package usecase

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/internal/domain/repository"
	"go.uber.org/zap"
)

type EpnbpUsecase interface {
	// GenerateNewPayUrl generates a new payment URL for a student bill
	GenerateNewPayUrl(user entity.User, mahasiswa entity.Mahasiswa, studentBill entity.StudentBill, budgetPeriod entity.BudgetPeriod) (*entity.PayUrl, error)
	
	// FindNotExpiredByStudentBill finds a payment URL that is not expired
	FindNotExpiredByStudentBill(studentBillID string) (*entity.PayUrl, error)
}

type epnbpUsecase struct {
	epnbpRepo repository.EpnbpRepository
	logger    *zap.Logger
}

func NewEpnbpUsecase(epnbpRepo repository.EpnbpRepository, logger *zap.Logger) EpnbpUsecase {
	return &epnbpUsecase{
		epnbpRepo: epnbpRepo,
		logger:    logger,
	}
}

func (u *epnbpUsecase) GenerateNewPayUrl(user entity.User, mahasiswa entity.Mahasiswa, studentBill entity.StudentBill, budgetPeriod entity.BudgetPeriod) (*entity.PayUrl, error) {
	// TODO: Implement business logic from backend/services/epnbp_service.go
	u.logger.Info("GenerateNewPayUrl", zap.Uint("studentBillID", studentBill.ID))
	return nil, nil
}

func (u *epnbpUsecase) FindNotExpiredByStudentBill(studentBillID string) (*entity.PayUrl, error) {
	return u.epnbpRepo.FindNotExpiredByStudentBill(studentBillID)
}




