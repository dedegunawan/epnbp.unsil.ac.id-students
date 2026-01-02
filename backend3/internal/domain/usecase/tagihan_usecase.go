package usecase

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend3/internal/domain/repository"
	"go.uber.org/zap"
)

type TagihanUsecase interface {
	// GetStudentBills gets all student bills for a student
	GetStudentBills(mhswID string, academicYear string) ([]entity.StudentBill, error)
	
	// GetActiveFinanceYearWithOverride gets active finance year with override
	GetActiveFinanceYearWithOverride(mahasiswa entity.Mahasiswa) (*entity.BudgetPeriod, error)
	
	// CreateNewTagihan creates a new tagihan for a student
	CreateNewTagihan(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod) error
	
	// CreateNewTagihanPasca creates a new tagihan for pascasarjana student
	CreateNewTagihanPasca(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod) error
	
	// DeleteUnpaidBills deletes unpaid bills for a student
	DeleteUnpaidBills(mhswID string, academicYear string) error
	
	// SavePaymentConfirmation saves a payment confirmation
	SavePaymentConfirmation(studentBill entity.StudentBill, vaNumber string, paymentDate string, objectName string) (*entity.PaymentConfirmation, error)
}

type tagihanUsecase struct {
	tagihanRepo repository.TagihanRepository
	logger      *zap.Logger
}

func NewTagihanUsecase(tagihanRepo repository.TagihanRepository, logger *zap.Logger) TagihanUsecase {
	return &tagihanUsecase{
		tagihanRepo: tagihanRepo,
		logger:      logger,
	}
}

func (u *tagihanUsecase) GetStudentBills(mhswID string, academicYear string) ([]entity.StudentBill, error) {
	return u.tagihanRepo.GetStudentBills(mhswID, academicYear)
}

func (u *tagihanUsecase) GetActiveFinanceYearWithOverride(mahasiswa entity.Mahasiswa) (*entity.BudgetPeriod, error) {
	return u.tagihanRepo.GetActiveFinanceYearWithOverride(mahasiswa)
}

func (u *tagihanUsecase) CreateNewTagihan(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod) error {
	// TODO: Implement business logic from backend/services/tagihan_service.go
	u.logger.Info("CreateNewTagihan", zap.String("mhswID", mahasiswa.MhswID))
	return nil
}

func (u *tagihanUsecase) CreateNewTagihanPasca(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod) error {
	// TODO: Implement business logic from backend/services/tagihan_service.go
	u.logger.Info("CreateNewTagihanPasca", zap.String("mhswID", mahasiswa.MhswID))
	return nil
}

func (u *tagihanUsecase) DeleteUnpaidBills(mhswID string, academicYear string) error {
	return u.tagihanRepo.DeleteUnpaidBills(mhswID, academicYear)
}

func (u *tagihanUsecase) SavePaymentConfirmation(studentBill entity.StudentBill, vaNumber string, paymentDate string, objectName string) (*entity.PaymentConfirmation, error) {
	// TODO: Implement business logic from backend/services/tagihan_service.go
	u.logger.Info("SavePaymentConfirmation", zap.Uint("studentBillID", studentBill.ID))
	return nil, nil
}




