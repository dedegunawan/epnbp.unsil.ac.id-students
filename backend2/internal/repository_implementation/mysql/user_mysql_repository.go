package mysql

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint64) (*entity.User, error) {
	var user entity.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindBySsoID(ssoID string) (*entity.User, error) {
	var user entity.User
	if err := r.db.Where("sso_id = ?", ssoID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(page, size int) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	r.db.Model(&entity.User{}).Count(&total)

	err := r.db.Offset((page - 1) * size).Limit(size).Find(&users).Error
	return users, total, err
}

func (r *UserRepository) UpdateAvatar(id uint64, url string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Update("avatar_url", url).Error
}

func (r *UserRepository) SetActive(id uint64, active bool) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Update("is_active", active).Error
}
