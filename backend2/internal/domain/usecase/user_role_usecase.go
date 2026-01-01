package usecase

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
)

type UserRoleUsecase interface {
	AssignRole(userID, roleID uint64) error
	RemoveRole(userID, roleID uint64) error
	GetRoles(userID uint64) ([]entity.Role, error)
}

type userRoleUsecase struct {
	userRoleService repository.UserRoleRepository
}

func NewUserRoleUsecase(userRoleService repository.UserRoleRepository) UserRoleUsecase {
	return &userRoleUsecase{userRoleService}
}

func (u *userRoleUsecase) AssignRole(userID, roleID uint64) error {
	return u.userRoleService.AssignRole(userID, roleID)
}

func (u *userRoleUsecase) RemoveRole(userID, roleID uint64) error {
	return u.userRoleService.RemoveRole(userID, roleID)
}

func (u *userRoleUsecase) GetRoles(userID uint64) ([]entity.Role, error) {
	return u.userRoleService.GetRoles(userID)
}
