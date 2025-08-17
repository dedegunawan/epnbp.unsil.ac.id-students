package entity

import (
	"time"
)

type RolePermission struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement;not null"`
	RoleID       uint64 `gorm:"not null"`
	PermissionID uint64 `gorm:"not null"`
	CreatedAt    time.Time
}
