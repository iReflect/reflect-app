package base_models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Feedback struct {
	gorm.Model
	FeedbackForm     FeedbackForm
	FeedbackFormID   uint `gorm:"not null"`
	ForUserProfile   UserProfile
	ForUserProfileID uint
	ByProfile        UserProfile
	ByUserProfileID  uint `gorm:"not null"`
	Team             Team
	TeamID           uint `gorm:"not null"`
	Status           int8 `gorm:"default:0; not null"` // TODO Add enum
	SubmittedAt      time.Time
	DurationStart    time.Time
	DurationEnd      time.Time
	ExpireAt         time.Time
}
