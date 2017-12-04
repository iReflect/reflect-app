package models

import "github.com/jinzhu/gorm"

// Category represent a category in feedback form
type Category struct {
	gorm.Model
	Title       string `gorm:"type:varchar(255); not null"`
	Description string `gorm:"type:text"`
}
