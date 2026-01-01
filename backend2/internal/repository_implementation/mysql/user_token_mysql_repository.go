package mysql

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
	g "gorm.io/gorm"
	"time"
)

type UserTokenRepository struct{ db *g.DB }

func NewUserTokenRepository(db *g.DB) repository.UserTokenRepository {
	return &UserTokenRepository{db}
}

func (r *UserTokenRepository) Create(token *entity.UserToken) error {
	return r.db.Create(token).Error
}

// Dapatkan token aktif terakhir user (bisa digunakan untuk refresh/logout)
func (r *UserTokenRepository) GetLatestByUserID(userID uint64) (*entity.UserToken, error) {
	var token entity.UserToken
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&token).Error
	return &token, err
}

// Dapatkan token aktif terakhir user (bisa digunakan untuk refresh/logout)
func (r *UserTokenRepository) GetByAccessToken(accessToken string) (*entity.UserToken, error) {
	var token entity.UserToken
	err := r.db.
		Where("access_token = ?", accessToken).
		First(&token).Error
	return &token, err
}

// Hapus semua token user (untuk logout semua sesi)
func (r *UserTokenRepository) DeleteByUserID(userID uint64) error {
	return r.db.Where("user_id = ?", userID).Delete(&entity.UserToken{}).Error
}

// Hapus token expired
func (r *UserTokenRepository) DeleteExpiredTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&entity.UserToken{}).Error
}
