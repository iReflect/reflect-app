package models

import "time"

// OTP ...
type OTP struct {
	Code       string `gorm:"type:varchar(16);not null"`
	ExpiryAt time.Time
	User       User
	UserID     uint `gorm:"unique"`
}
