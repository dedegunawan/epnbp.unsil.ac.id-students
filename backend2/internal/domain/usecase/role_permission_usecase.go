package usecase

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
)

type RolePermissionUsecase interface {
	AssignPermission(roleID, permissionID uint64) error
	RemovePermission(roleID, permissionID uint64) error
}

type rolePermissionUsecase struct {
	rolePermissionService repository.RolePermissionRepository
}

func NewRolePermissionUsecase(rolePermissionService repository.RolePermissionRepository) RolePermissionUsecase {
	return &rolePermissionUsecase{rolePermissionService}
}

func (u *rolePermissionUsecase) AssignPermission(roleID, permissionID uint64) error {
	return u.rolePermissionService.AssignPermission(roleID, permissionID)
}

func (u *rolePermissionUsecase) RemovePermission(roleID, permissionID uint64) error {
	return u.rolePermissionService.RemovePermission(roleID, permissionID)
}
