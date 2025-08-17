package entity

import (
	"time"
)

type UserRole struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement;not null"`
	UserID    uint64 `gorm:"not null"`
	RoleID    uint64 `gorm:"not null"`
	CreatedAt time.Time
}
