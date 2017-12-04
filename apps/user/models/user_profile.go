package models

import "github.com/jinzhu/gorm"

type UserProfile struct {
	gorm.Model
	User    User
	UserId  uint `gorm:"not null"`
	Role    Role
	RoleID  uint `gorm:"not null"`
	Active bool `gorm:"default:true; not null"`
}
