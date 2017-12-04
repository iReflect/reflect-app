package models

import "github.com/jinzhu/gorm"

// Team represent a team/project comprising a set of user
type Team struct {
	gorm.Model
	Name        string `gorm:"type:varchar(64);not null"`
	Description string `gorm:"type:text"`
	Active      bool   `gorm:"default:true; not null"`
	Users       []User
}
