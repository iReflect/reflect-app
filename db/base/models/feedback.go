package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Feedback represent a submitted/in-progress feedback form by a user
type Feedback struct {
	gorm.Model
	FeedbackForm     FeedbackForm
	Title            string `gorm:"type:varchar(255); not null"`
	FeedbackFormID   uint   `gorm:"not null"`
	ForUserProfile   UserProfile
	ForUserProfileID uint
	ByProfile        UserProfile
	ByUserProfileID  uint   `gorm:"not null"`
	Team             Team
	TeamID           uint   `gorm:"not null"`
	Status           int8   `gorm:"default:0; not null"` // TODO Add enum
	SubmittedAt      time.Time
	DurationStart    time.Time
	DurationEnd      time.Time
	ExpireAt         time.Time
}
