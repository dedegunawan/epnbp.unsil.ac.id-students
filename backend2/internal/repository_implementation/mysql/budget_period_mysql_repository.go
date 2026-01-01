package mysql

import (
	"fmt"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
	"gorm.io/gorm"
)

type BudgetPeriodRepository struct {
	db     *gorm.DB
	logger *logger.Logger
}

func NewBudgetPeriodRepository(db *gorm.DB, logger *logger.Logger) repository.BudgetPeriodRepository {
	return &BudgetPeriodRepository{db: db, logger: logger}
}

// Create inserts a new record into the database.
func (r BudgetPeriodRepository) Create(entity *entity.BudgetPeriod) error {
	return r.db.Create(entity).Error
}

// FindByID retrieves a record by its ID.
func (r BudgetPeriodRepository) FindByID(id uint64) (*entity.BudgetPeriod, error) {
	var budget entity.BudgetPeriod
	err := r.db.Where("id = ?", id).First(&budget).Error
	return &budget, err
}

func (r BudgetPeriodRepository) IsKnownField(field string) (bool, error) {
	allowedFields := map[string]bool{
		"id":                 true,
		"kode":               true,
		"year":               true,
		"semester":           true,
		"name":               true,
		"start_date":         true,
		"end_date":           true,
		"fiscal_year":        true,
		"payment_start_date": true,
		"payment_end_date":   true,
		"is_active":          true,
		"description":        true,
		"created_at":         true,
		"updated_at":         true,
		"deleted_at":         true,
	}

	if !allowedFields[field] {
		return false, fmt.Errorf("invalid field: %s", field)
	}
	return true, nil
}

// FindBy any field retrieves a record by a specific field and value.
func (r BudgetPeriodRepository) FindBy(field string, value any) (*entity.BudgetPeriod, error) {

	_, err := r.IsKnownField(field)
	if err != nil {
		return nil, err
	}

	var budget entity.BudgetPeriod
	err = r.db.Where(fmt.Sprintf("%s = ?", field), value).First(&budget).Error
	return &budget, err
}

// List retrieves a list of records with pagination.
func (r BudgetPeriodRepository) List(page, size int) ([]entity.BudgetPeriod, int64, error) {
	var users []entity.BudgetPeriod
	var total int64

	r.db.Model(&entity.BudgetPeriod{}).Count(&total)

	err := r.db.Offset((page - 1) * size).Limit(size).Find(&users).Error
	return users, total, err

}

// Update modifies an existing record.
func (r BudgetPeriodRepository) Update(entity *entity.BudgetPeriod) error {
	return r.db.Save(entity).Error
}

// Delete removes a record by its ID.
func (r BudgetPeriodRepository) Delete(id uint64) error {
	return r.db.Where("id = ?", id).Delete(entity.BudgetPeriod{}).Error
}

// Count returns the total number of records for a given entity type.
func (r BudgetPeriodRepository) Count(filterInterface interface{}) (int64, error) {
	var count int64
	err := r.db.Model(&entity.BudgetPeriod{}).Where(filterInterface).Count(&count).Error
	if err != nil {
		r.logger.Error("Failed to count budget periods", "error", err)
		return 0, err
	}
	return count, nil
}

// GetAll retrieves all records of a specific type.
func (r BudgetPeriodRepository) GetAll() ([]entity.BudgetPeriod, error) {
	var budgetPeriod []entity.BudgetPeriod
	if err := r.db.Find(&budgetPeriod).Error; err != nil {
		r.logger.Error("Failed to retrieve all budget periods", "error", err)
		return nil, err
	}
	return budgetPeriod, nil
}

// GetBy any field retrieves a record by a specific field and value.
func (r BudgetPeriodRepository) GetBy(filterInterface interface{}) ([]entity.BudgetPeriod, error) {
	var budgetPeriod []entity.BudgetPeriod
	err := r.db.Where(filterInterface).Find(&budgetPeriod).Error
	if err != nil {
		r.logger.Error("Failed to retrieve budget periods by filter", "error", err)
		return nil, err
	}
	return budgetPeriod, nil
}

func (r BudgetPeriodRepository) SetActive(entity *entity.BudgetPeriod) error {
	entity.IsActive = true
	return r.db.Save(entity).Error
}
func (r BudgetPeriodRepository) GetActive() (*entity.BudgetPeriod, error) {
	return r.FindBy("is_active", true)
}
