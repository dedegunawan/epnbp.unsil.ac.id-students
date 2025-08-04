package models

import "gorm.io/gorm"

type SintesysCallback struct {
	ID       uint   `gorm:"primaryKey"`
	Url      string `gorm:"column:url" json:"url"`
	Data     string `gorm:"column:data" json:"data"`
	Response string `gorm:"column:response" json:"response"`
}

func MigrateSintesys(db *gorm.DB) {
	db.AutoMigrate(&SintesysCallback{})
}
