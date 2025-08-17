package mysql

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/user_token"
	g "gorm.io/gorm"
	"time"
)

type userTokenRepository struct{ db *g.DB }

func NewUserTokenRepository(db *g.DB) user_token.Repository {
	_ = db.AutoMigrate(&user_token.UserToken{})
	return &userTokenRepository{db}
}

func (r *userTokenRepository) Create(token *user_token.UserToken) error {
	return r.db.Create(token).Error
}

// Dapatkan token aktif terakhir user (bisa digunakan untuk refresh/logout)
func (r *userTokenRepository) FindLatestByUserID(userID uint64) (*user_token.UserToken, error) {
	var token user_token.UserToken
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&token).Error
	return &token, err
}

// Dapatkan token aktif terakhir user (bisa digunakan untuk refresh/logout)
func (r *userTokenRepository) FindByAccessToken(accessToken string) (*user_token.UserToken, error) {
	var token user_token.UserToken
	err := r.db.
		Where("access_token = ?", accessToken).
		First(&token).Error
	return &token, err
}

// Hapus semua token user (untuk logout semua sesi)
func (r *userTokenRepository) DeleteByUserID(userID uint64) error {
	return r.db.Where("user_id = ?", userID).Delete(&user_token.UserToken{}).Error
}

// Hapus token expired
func (r *userTokenRepository) DeleteExpiredTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&user_token.UserToken{}).Error
}
