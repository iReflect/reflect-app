package models

import "github.com/jinzhu/gorm"

/*
 * TODO Add support for versioning
 * TODO Add support for marking row non editable
 */
type Skill struct {
	gorm.Model
	Title        string `gorm:"type:varchar(255); not null"`
	DisplayTitle string `gorm:"type:varchar(255)"`
	Description  string `gorm:"type:text"`
	Weight       int    `gorm:"default:1"`
	Questions    []Question
}
