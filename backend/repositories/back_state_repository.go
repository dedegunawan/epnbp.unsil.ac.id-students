package repositories

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"gorm.io/gorm"
)

type BackState interface {
	FindById(id uint) (*BackState, error)
}

type backStateRepository struct {
	DB *gorm.DB
}

func NewBackStateRepository(db *gorm.DB) BackState {
	return &backStateRepository{DB: database.DB}
}

func (bs *backStateRepository) FindById(id uint) (*BackState, error) {
	var backState BackState
	err := bs.DB.Where("id", id).First(&backState).Error
	return &backState, err
}
