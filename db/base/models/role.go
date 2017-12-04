package models

import "github.com/jinzhu/gorm"

// Role represent supported user roles
type Role struct {
	gorm.Model
	Name string `gorm:"type:varchar(64)"`
}
