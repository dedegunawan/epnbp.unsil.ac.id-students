package entity

import (
	"gorm.io/gorm"
	"time"
)

type Role struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement;not null" json:"id"`
	Name        string         `gorm:"type:varchar(100);unique;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Users       []User         `gorm:"many2many:user_roles;" json:"-"`
	Permissions []Permission   `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
