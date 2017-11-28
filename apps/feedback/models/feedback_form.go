package models

import "github.com/jinzhu/gorm"

/*
 * TODO Add support for versioning
 */
type FeedbackForm struct {
	gorm.Model
	Title       string `gorm:"type:text; not null"`
	Description string `gorm:"type:text;"`
	Status      int8   `gorm:"default:0; not null"`
	Archive     bool   `gorm:"default:false; not null"`
}
