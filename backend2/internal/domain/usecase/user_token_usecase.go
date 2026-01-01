package usecase

import (
	"context"
	"encoding/json"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/jwtmanager"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
	"time"
)

type UserTokenUsecase interface {
	SaveUserToken(userID uint64, jwtType, accessToken, refreshToken, tokenType string, expiresAt time.Time) error
	SaveLoginUserToken(userID uint64, accessToken string) (*entity.UserToken, error)
	GetLatestToken(userID uint64) (*entity.UserToken, error)
	CleanupExpiredTokens() error
	GetContext() context.Context
	SetContext(ctx context.Context)
	GetByAccessToken(token string) (*entity.UserToken, error)
}

type userTokenUsecase struct {
	userTokenRepository repository.UserTokenRepository
	context             context.Context
	logger              *logger.Logger
	jwt                 *jwtmanager.Manager
}

func NewUserTokenUsecase(userTokenRepository repository.UserTokenRepository, ctx context.Context, lg *logger.Logger, jwt *jwtmanager.Manager) UserTokenUsecase {
	return &userTokenUsecase{
		userTokenRepository: userTokenRepository,
		context:             ctx,
		logger:              lg,
		jwt:                 jwt,
	}

}

// Simpan token dari login OIDC
func (s *userTokenUsecase) SaveUserToken(userID uint64, jwtType, accessToken, refreshToken, tokenType string, expiresAt time.Time) error {
	token := &entity.UserToken{
		UserID:       userID,
		JwtType:      jwtType,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}
	json_, err := json.Marshal(&token)
	s.logger.Info(userID, "json token : ", string(json_), err)
	return s.userTokenRepository.Create(token)
}

// Simpan token dari login OIDC
func (s *userTokenUsecase) SaveLoginUserToken(userID uint64, accessToken string) (*entity.UserToken, error) {
	refreshToken, err := s.jwt.Generate(userID, "refresh", "refresh")
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

	token := &entity.UserToken{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
		JwtType:      entity.JWTTypeInternal,
		Fingerprint:  fingerprint,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
	}
	err = s.userTokenRepository.Create(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Ambil token terbaru (misalnya untuk logout)
func (s *userTokenUsecase) GetLatestToken(userID uint64) (*entity.UserToken, error) {
	return s.userTokenRepository.GetLatestByUserID(userID)
}

func (s *userTokenUsecase) GetByAccessToken(token string) (*entity.UserToken, error) {
	return s.userTokenRepository.GetByAccessToken(token)
}

// Bersihkan token expired
func (s *userTokenUsecase) CleanupExpiredTokens() error {
	return s.userTokenRepository.DeleteExpiredTokens()
}

func (s *userTokenUsecase) GetContext() context.Context {
	return s.context
}

func (s *userTokenUsecase) SetContext(ctx context.Context) {
	s.context = ctx
}

func (s *userTokenUsecase) getFingerprintFromContext() string {
	// misal ambil dari context middleware
	if v, ok := s.context.Value("fingerprint").(string); ok {
		return v
	}
	return ""
}

func (s *userTokenUsecase) getUserAgentFromContext() string {
	if v, ok := s.context.Value("user_agent").(string); ok {
		return v
	}
	return ""
}

func (s *userTokenUsecase) getIPAddressFromContext() string {
	if v, ok := s.context.Value("ip_address").(string); ok {
		return v
	}
	return ""
}
