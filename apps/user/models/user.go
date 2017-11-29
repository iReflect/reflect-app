package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Email     string `gorm:"type:varchar(255); not null; unique_index"`
	FirstName string `gorm:"type:varchar(30); not null"`
	LastName  string `gorm:"type:varchar(30)"`
	Role      Role
	RoleID    uint   `gorm:"not null"`
	Active    bool   `gorm:"default:true; not null"`
	Teams     []Team `gorm:"many2many:user_team_associations;"`
}
