package entity

import (
	"time"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;not null"`
	Name         string    `gorm:"column:name"`
	Email        string    `json:"email" gorm:"size:180;uniqueIndex;not null"`
	PasswordHash *string   `json:"-" gorm:"size:255;default:null"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	AvatarURL    string    `json:"avatar,omitempty" gorm:"size:255"`
	SsoID        *string   `json:"sso_id,omitempty" gorm:"size:255;default:null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
