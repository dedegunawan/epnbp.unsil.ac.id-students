package repository

import "github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"

type RoleRepository interface {
	Create(role *entity.Role) error
	Update(role *entity.Role) error
	Delete(roleID uint64) error
	FindByID(id uint64) (*entity.Role, error)
	GetAll() ([]entity.Role, error)
}
