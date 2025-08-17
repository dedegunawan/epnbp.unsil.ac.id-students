package usecase

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/service"
)

type UserUsecase interface {
	Register(name, email, password string) (*entity.User, error)
	GetByID(id uint64) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	List(page, size int) ([]entity.User, int64, error)
	UpdateAvatar(id uint64, url string) error
	SetActive(id uint64, active bool) error
}

type userUsecase struct {
	userService service.UserService
}

func NewUserUsecase(userSvc service.UserService) UserUsecase {
	return &userUsecase{userService: userSvc}
}

// Implement each method by calling the corresponding service...
