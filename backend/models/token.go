package models

import (
	"github.com/google/uuid"
	"time"
)

const (
	JWTTypeKeycloak = "keycloak"
	JWTTypeInternal = "internal"
)

type UserToken struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;index"`
	AccessToken  string    `gorm:"type:text"`
	RefreshToken string    `gorm:"type:text"`
	ExpiresAt    time.Time
	TokenType    string  `gorm:"type:varchar(20)"`
	JwtType      string  `gorm:"type:jwt_type_enum;default:'keycloak'"` // pakai enum dari DB
	Fingerprint  string  `gorm:"type:text"`
	UserAgent    string  `gorm:"type:text"`
	IPAddress    *string `gorm:"type:inet"`

	CreatedAt time.Time
}
