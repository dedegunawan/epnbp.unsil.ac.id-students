package entity

import (
	"time"
)

const (
	JWTTypeKeycloak = "keycloak"
	JWTTypeInternal = "internal"
)

type UserToken struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement;not null"`
	UserID       uint64 `gorm:"column:user_id"`
	AccessToken  string `gorm:"type:text"`
	RefreshToken string `gorm:"type:text"`
	ExpiresAt    time.Time
	TokenType    string  `gorm:"type:varchar(20)"`
	JwtType      string  `gorm:"type:jwt_type_enum;default:'keycloak'"` // pakai enum dari DB
	Fingerprint  string  `gorm:"type:text"`
	UserAgent    string  `gorm:"type:text"`
	IPAddress    *string `gorm:"type:inet"`
	CreatedAt    time.Time
}
