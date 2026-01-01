package usecase

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
)

type RoleUsecase interface {
	Create(role *entity.Role) error
	Update(role *entity.Role) error
	Delete(roleID uint64) error
	FindByID(roleID uint64) (*entity.Role, error)
	GetAll() ([]entity.Role, error)
}

type roleUsecase struct {
	roleService repository.RoleRepository
}

func NewRoleUsecase(roleService repository.RoleRepository) RoleUsecase {
	return &roleUsecase{roleService}
}

func (r *roleUsecase) Create(role *entity.Role) error {
	return r.roleService.Create(role)
}

func (r *roleUsecase) Update(role *entity.Role) error {
	return r.roleService.Update(role)
}

func (r *roleUsecase) Delete(roleID uint64) error {
	return r.roleService.Delete(roleID)
}

func (r *roleUsecase) FindByID(id uint64) (*entity.Role, error) {
	return r.roleService.FindByID(id)
}

func (r *roleUsecase) GetAll() ([]entity.Role, error) {
	return r.roleService.GetAll()
}
