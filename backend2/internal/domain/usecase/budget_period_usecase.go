package usecase

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
)

type BudgetPeriodUsecase interface {
	// Create inserts a new record into the database.
	Create(entity *entity.BudgetPeriod) error
	// FindByID retrieves a record by its ID.
	FindByID(id uint64) (*entity.BudgetPeriod, error)
	// FindBy any field retrieves a record by a specific field and value.
	FindBy(field string, value any) (*entity.BudgetPeriod, error)

	// List retrieves a list of records with pagination.
	List(page, size int) ([]entity.BudgetPeriod, int64, error)
	// Update modifies an existing record.
	Update(entity *entity.BudgetPeriod) error
	// Delete removes a record by its ID.
	Delete(id uint64) error
	// Count returns the total number of records for a given entity type.
	Count(filterInterface interface{}) (int64, error)
	// GetAll retrieves all records of a specific type.
	GetAll() ([]entity.BudgetPeriod, error)
	// GetBy any field retrieves a record by a specific field and value.
	GetBy(filterInterface interface{}) ([]entity.BudgetPeriod, error)

	SetActive(entity *entity.BudgetPeriod) error
	GetActive() (*entity.BudgetPeriod, error)
}

type budgetPeriodUsecase struct {
	budgetPeriodRepository repository.BudgetPeriodRepository
	logger                 *logger.Logger
}

func NewBudgetPeriodUsecase(budgetPeriodRepository repository.BudgetPeriodRepository, logger *logger.Logger) BudgetPeriodUsecase {
	return &budgetPeriodUsecase{budgetPeriodRepository: budgetPeriodRepository, logger: logger}
}

// Create inserts a new record into the database.
func (usecase budgetPeriodUsecase) Create(entity *entity.BudgetPeriod) error {
	return usecase.budgetPeriodRepository.Create(entity)
}

// FindByID retrieves a record by its ID.
func (usecase budgetPeriodUsecase) FindByID(id uint64) (*entity.BudgetPeriod, error) {
	return usecase.budgetPeriodRepository.FindByID(id)
}

// FindBy any field retrieves a record by a specific field and value.
func (usecase budgetPeriodUsecase) FindBy(field string, value any) (*entity.BudgetPeriod, error) {
	return usecase.budgetPeriodRepository.FindBy(field, value)
}

// List retrieves a list of records with pagination.
func (usecase budgetPeriodUsecase) List(page, size int) ([]entity.BudgetPeriod, int64, error) {
	return usecase.budgetPeriodRepository.List(page, size)
}

// Update modifies an existing record.
func (usecase budgetPeriodUsecase) Update(entity *entity.BudgetPeriod) error {
	return usecase.budgetPeriodRepository.Update(entity)
}

// Delete removes a record by its ID.
func (usecase budgetPeriodUsecase) Delete(id uint64) error {
	return usecase.budgetPeriodRepository.Delete(id)
}

// Count returns the total number of records for a given entity type.
func (usecase budgetPeriodUsecase) Count(filterInterface interface{}) (int64, error) {
	return usecase.budgetPeriodRepository.Count(filterInterface)
}

// GetAll retrieves all records of a specific type.
func (usecase budgetPeriodUsecase) GetAll() ([]entity.BudgetPeriod, error) {
	return usecase.budgetPeriodRepository.GetAll()
}

// GetBy any field retrieves a record by a specific field and value.
func (usecase budgetPeriodUsecase) GetBy(filterInterface interface{}) ([]entity.BudgetPeriod, error) {
	return usecase.budgetPeriodRepository.GetBy(filterInterface)
}

func (usecase budgetPeriodUsecase) SetActive(entity *entity.BudgetPeriod) error {
	return usecase.budgetPeriodRepository.SetActive(entity)
}
func (usecase budgetPeriodUsecase) GetActive() (*entity.BudgetPeriod, error) {
	return usecase.budgetPeriodRepository.GetActive()
}
