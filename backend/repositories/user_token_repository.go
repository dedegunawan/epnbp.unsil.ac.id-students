package repositories

import (
	"time"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserTokenRepository struct {
	DB *gorm.DB
}

// Simpan token baru (biasanya dipanggil saat login)
func (r *UserTokenRepository) Create(token *models.UserToken) error {
	return r.DB.Create(token).Error
}

// Dapatkan token aktif terakhir user (bisa digunakan untuk refresh/logout)
func (r *UserTokenRepository) FindLatestByUserID(userID uuid.UUID) (*models.UserToken, error) {
	var token models.UserToken
	err := r.DB.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&token).Error
	return &token, err
}

// Dapatkan token aktif terakhir user (bisa digunakan untuk refresh/logout)
func (r *UserTokenRepository) FindByAccessToken(accessToken string) (*models.UserToken, error) {
	var token models.UserToken
	err := r.DB.
		Where("access_token = ?", accessToken).
		First(&token).Error
	return &token, err
}

// Hapus semua token user (untuk logout semua sesi)
func (r *UserTokenRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.DB.Where("user_id = ?", userID).Delete(&models.UserToken{}).Error
}

// Hapus token expired
func (r *UserTokenRepository) DeleteExpiredTokens() error {
	return r.DB.Where("expires_at < ?", time.Now()).Delete(&models.UserToken{}).Error
}
