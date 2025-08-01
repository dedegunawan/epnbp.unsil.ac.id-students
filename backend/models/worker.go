package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type JobQueue struct {
	ID         uint           `gorm:"primaryKey"`
	Type       string         `gorm:"type:text"`
	Payload    datatypes.JSON `gorm:"type:jsonb"`
	Status     string         `gorm:"type:text"`
	Retries    int
	MaxRetries int
	RunAt      time.Time
	LastError  *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func MigrateJob(db *gorm.DB) {
	db.AutoMigrate(&JobQueue{})

}
