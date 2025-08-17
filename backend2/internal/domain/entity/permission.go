package entity

import (
	"gorm.io/gorm"
	"time"
)

type Permission struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement;not null" json:"id"`
	Name        string         `gorm:"type:varchar(150);unique;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Roles       []Role         `gorm:"many2many:role_permissions;" json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
