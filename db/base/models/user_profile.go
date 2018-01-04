package models

import "github.com/jinzhu/gorm"

// UserProfile associated to app users
// A user can have multiple profile, but only one of them could be active at any moment
type UserProfile struct {
	gorm.Model
	User   User
	UserID uint `gorm:"not null"`
	Role   Role
	RoleID uint `gorm:"not null"`
	Active bool `gorm:"default:false; not null"`
}
