package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Email string `gorm:"type:varchar(100);unique_index"`
	Name  string
	Type  string
}

type Group struct {
	gorm.Model
	Name string
}

type Role struct {
	gorm.Model
	Name    string
	UserID  uint `gorm:"index"`
	GroupID uint `gorm:"index"`
	Type    string
}
