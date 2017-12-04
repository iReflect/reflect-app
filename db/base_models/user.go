package base_models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Email     string `gorm:"type:varchar(255); not null; unique_index"`
	FirstName string `gorm:"type:varchar(30); not null"`
	LastName  string `gorm:"type:varchar(150)"`
	Active    bool   `gorm:"default:true; not null"`
	Teams     []Team
	Profiles  []UserProfile
}
