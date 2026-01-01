package repository

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
)

type UserTokenRepository interface {
	Create(token *entity.UserToken) error
	GetLatestByUserID(userID uint64) (*entity.UserToken, error)
	GetByAccessToken(accessToken string) (*entity.UserToken, error)
	DeleteByUserID(userID uint64) error
	DeleteExpiredTokens() error
}
