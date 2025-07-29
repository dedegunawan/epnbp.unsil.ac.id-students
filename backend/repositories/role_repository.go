package repositories

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"gorm.io/gorm"
)

type RoleRepository struct {
	DB *gorm.DB
}

func (r *RoleRepository) FindByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.DB.Preload("Permissions").Where("name = ?", name).First(&role).Error
	return &role, err
}

func (r *RoleRepository) GetAll() ([]models.Role, error) {
	var roles []models.Role
	err := r.DB.Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) Create(role *models.Role) error {
	return r.DB.Create(role).Error
}
