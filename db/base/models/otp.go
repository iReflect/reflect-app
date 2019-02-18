package models

import (
	"github.com/jinzhu/gorm"
)

// OTP ...
type OTP struct {
	gorm.Model
	Code   string `gorm:"type:varchar(16);not null"`
	User   User
	UserID uint `gorm:"not null"`
}
