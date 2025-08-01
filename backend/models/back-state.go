package models

import (
	"gorm.io/gorm"
	"time"
)

type BackState struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BackState string    `gorm:"back_state" json:"back_state"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func MigrateBackState(db *gorm.DB) {
	db.AutoMigrate(
		&BackState{},
	)
}
