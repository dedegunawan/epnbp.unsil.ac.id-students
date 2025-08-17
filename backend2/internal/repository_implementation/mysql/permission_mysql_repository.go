package mysql

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"gorm.io/gorm"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db}
}

func (r *PermissionRepository) GetAll() ([]entity.Permission, error) {
	var permissions []entity.Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}

func (r *PermissionRepository) GetByRoleID(roleID uint64) ([]entity.Permission, error) {
	var permissions []entity.Permission
	err := r.db.Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).Find(&permissions).Error
	return permissions, err
}
