package usecase

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
)

type PermissionUsecase interface {
	GetAll() ([]entity.Permission, error)
	GetByRoleID(roleID uint64) ([]entity.Permission, error)
}

type permissionUsecase struct {
	permissionService repository.PermissionRepository
}

func NewPermissionUsecase(permissionService repository.PermissionRepository) PermissionUsecase {
	return &permissionUsecase{permissionService}
}

func (p *permissionUsecase) GetAll() ([]entity.Permission, error) {
	return p.permissionService.GetAll()
}

func (p *permissionUsecase) GetByRoleID(roleID uint64) ([]entity.Permission, error) {
	return p.permissionService.GetByRoleID(roleID)
}
