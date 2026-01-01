package repository

import "github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"

type BudgetPeriodRepository interface {
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
	GetBy(anyInterface interface{}) ([]entity.BudgetPeriod, error)

	SetActive(entity *entity.BudgetPeriod) error
	GetActive() (*entity.BudgetPeriod, error)
}
