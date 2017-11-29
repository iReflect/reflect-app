package models

import "github.com/jinzhu/gorm"

type Team struct {
	gorm.Model
	Name        string `gorm:"type:varchar(64);not null"`
	Description string `gorm:"type:text"`
	Active      bool   `gorm:"default:true; not null"`
	Users       []User `gorm:"many2many:user_team_associations;"`
}
