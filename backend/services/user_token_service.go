package services

import (
	"context"
	"github.com/dedegunawan/backend-ujian-telp-v5/config"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"time"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/google/uuid"
)

type UserTokenService struct {
	Repo    *repositories.UserTokenRepository
	Context context.Context
}

// Simpan token dari login OIDC
func (s *UserTokenService) SaveUserToken(userID uuid.UUID, accessToken, refreshToken, tokenType string, expiresAt time.Time) error {
	token := &models.UserToken{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}
	return s.Repo.Create(token)
}

// Simpan token dari login OIDC
func (s *UserTokenService) SaveLoginUserToken(userID uuid.UUID, accessToken string) (*models.UserToken, error) {
	refreshToken, err := utils.GenerateJWT(userID, "refresh", "refresh", config.GetDefaultRefreshTokenExpired())
	if err != nil {
		refreshToken = accessToken + "refresh"
	}

	tokenType := "Bearer"
	expiresAt := time.Now().Add(1 * time.Hour)

	// Ambil data tambahan dari internal context/service
	fingerprint := s.getFingerprintFromContext()
	userAgent := s.getUserAgentFromContext()
	var ipAddress *string
	ip := s.getIPAddressFromContext()
	if ip != "" {
		ipAddress = &ip
	} else {
		ipAddress = nil
	}

	token := &models.UserToken{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
		JwtType:      models.JWTTypeInternal,
		Fingerprint:  fingerprint,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
	}
	err = s.Repo.Create(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Ambil token terbaru (misalnya untuk logout)
func (s *UserTokenService) GetLatestToken(userID uuid.UUID) (*models.UserToken, error) {
	return s.Repo.FindLatestByUserID(userID)
}

// Bersihkan token expired
func (s *UserTokenService) CleanupExpiredTokens() error {
	return s.Repo.DeleteExpiredTokens()
}

func (s *UserTokenService) SetContext(ctx context.Context) {
	s.Context = ctx
}

func (s *UserTokenService) getFingerprintFromContext() string {
	// misal ambil dari context middleware
	if v, ok := s.Context.Value("fingerprint").(string); ok {
		return v
	}
	return ""
}

func (s *UserTokenService) getUserAgentFromContext() string {
	if v, ok := s.Context.Value("user_agent").(string); ok {
		return v
	}
	return ""
}

func (s *UserTokenService) getIPAddressFromContext() string {
	if v, ok := s.Context.Value("ip_address").(string); ok {
		return v
	}
	return ""
}
