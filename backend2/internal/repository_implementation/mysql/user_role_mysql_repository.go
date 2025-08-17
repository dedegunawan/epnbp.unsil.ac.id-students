package mysql

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
	"gorm.io/gorm"
)

type UserRoleRepository struct {
	db *gorm.DB
}

func NewUserRoleRepository(db *gorm.DB) repository.PermissionRepository {
	return &UserRoleRepository{db}
}

func (r *UserRoleRepository) AssignRole(userID, roleID uint64) error {
	userRole := &entity.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.Create(userRole).Error
}

func (r *UserRoleRepository) RemoveRole(userID, roleID uint64) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&entity.UserRole{}).Error
}

func (r *UserRoleRepository) GetRolesByUserID(userID uint64) ([]entity.Role, error) {
	var roles []entity.Role
	err := r.db.Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).Find(&roles).Error
	return roles, err
}
