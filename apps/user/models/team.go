package models

import "github.com/jinzhu/gorm"

type Team struct {
	gorm.Model
	Name        string `gorm:"type:varchar(30);not null"`
	Description string `gorm:"type:varchar"`
	IsActive    bool   `gorm:"default:true; not null"`
}
