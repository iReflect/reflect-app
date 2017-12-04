package models

import "github.com/jinzhu/gorm"

// Skill represent the skill comprised by category
// TODO Add support for versioning and making it non-editable
type Skill struct {
	gorm.Model
	Title        string `gorm:"type:varchar(255); not null"`
	DisplayTitle string `gorm:"type:varchar(255)"`
	Description  string `gorm:"type:text"`
	Weight       int    `gorm:"default:1"`
	Questions    []Question
}
