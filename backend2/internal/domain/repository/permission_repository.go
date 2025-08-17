package repository

import "github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"

type PermissionRepository interface {
	GetAll() ([]entity.Permission, error)
	GetByRoleID(roleID uint64) ([]entity.Permission, error)
}
