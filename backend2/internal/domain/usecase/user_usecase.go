package usecase

import (
	"errors"
	"fmt"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/pointer"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	Register(name, email, password string) (*entity.User, error)
	GetByID(id uint64) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	List(page, size int) ([]entity.User, int64, error)
	UpdateAvatar(id uint64, url string) error
	SetActive(id uint64, active bool) error
	GetOrCreateByEmail(ssoID string, email string, name string) (*entity.User, error)
}

type userUsecase struct {
	userService repository.UserRepository
}

func NewUserUsecase(userSvc repository.UserRepository) UserUsecase {
	return &userUsecase{userService: userSvc}
}

// Implement each method by calling the corresponding service...
func (u *userUsecase) Register(name, email, password string) (*entity.User, error) {
	if name == "" || email == "" || password == "" {
		return nil, errors.New("name, email, and password must not be empty")
	}

	existingUser, _ := u.userService.FindByEmail(email)
	if existingUser != nil {
		return nil, errors.New("email is already registered")
	}

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entity.User{
		Name:         name,
		Email:        email,
		PasswordHash: pointer.Of(string(passwordHashed)),
	}

	err = u.userService.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil

}

func (u *userUsecase) GetByID(id uint64) (*entity.User, error) {
	return u.userService.FindByID(id)
}

func (u *userUsecase) GetByEmail(email string) (*entity.User, error) {
	return u.userService.FindByEmail(email)
}

func (u *userUsecase) List(page, size int) ([]entity.User, int64, error) {
	return u.userService.List(page, size)
}

func (u *userUsecase) UpdateAvatar(id uint64, url string) error {
	return u.userService.UpdateAvatar(id, url)
}

func (u *userUsecase) SetActive(id uint64, active bool) error {
	return u.userService.SetActive(id, active)
}

func (r *userUsecase) GetOrCreateByEmail(ssoID string, email string, name string) (*entity.User, error) {
	if ssoID == "" || email == "" {
		return nil, errors.New("SSO ID and email must not be empty")
	}
	
	user, err := r.userService.FindByEmail(email)
	if user != nil && err == nil {
		return user, nil
	}

	user = &entity.User{
		Name:     name,
		Email:    email,
		SsoID:    &ssoID,
		IsActive: true,
	}
	if err := r.userService.Create(user); err != nil {
		return nil, err
	}
	return user, nil

}
