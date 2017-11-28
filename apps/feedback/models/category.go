package models

import "github.com/jinzhu/gorm"

type Category struct {
	gorm.Model
	Title       string `gorm:"type:varchar; not null"`
	Description string `gorm:"type:text"`
}
