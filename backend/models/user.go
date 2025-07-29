// models/user.go
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100)" json:"name"`
	Email     string         `gorm:"type:varchar(150);unique;not null" json:"email"`
	Password  *string        `gorm:"type:text" json:"password,omitempty"`
	SSOID     *string        `gorm:"type:varchar(255);index" json:"sso_id,omitempty"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	Roles     []Role         `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100);unique;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Users       []User         `gorm:"many2many:user_roles;" json:"-"`
	Permissions []Permission   `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserRole struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time
}

type Permission struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(150);unique;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Roles       []Role         `gorm:"many2many:role_permissions;" json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type RolePermission struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RoleID       uuid.UUID `gorm:"type:uuid;not null"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt    time.Time
}
