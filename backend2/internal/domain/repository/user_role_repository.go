package repository

import "github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"

type UserRoleRepository interface {
	AssignRole(userID, roleID uint64) error
	RemoveRole(userID, roleID uint64) error
	GetRoles(userID uint64) ([]entity.Role, error)
}
