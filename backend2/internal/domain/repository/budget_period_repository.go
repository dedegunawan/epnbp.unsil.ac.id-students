package repository

import "github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"

type UserRepository interface {
	Create(user *entity.User) error
	FindByID(id uint64) (*entity.User, error)
	FindByEmail(email string) (*entity.User, error)
	List(page, size int) ([]entity.User, int64, error)
	UpdateAvatar(id uint64, url string) error
	SetActive(id uint64, active bool) error
}
