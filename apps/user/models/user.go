package models

import "github.com/jinzhu/gorm"

// User represent the app user in system
type User struct {
	gorm.Model
	Email     string `gorm:"type:varchar(255); not null; unique_index"`
	FirstName string `gorm:"type:varchar(30); not null"`
	LastName  string `gorm:"type:varchar(150)"`
	Active    bool   `gorm:"default:true; not null"`
	Teams     []Team
	Profiles  []UserProfile
}
