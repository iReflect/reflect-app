package models

import (
	"time"

	"github.com/jinzhu/gorm"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

// Feedback represent a submitted/in-progress feedback form by a user
type Feedback struct {
	gorm.Model
	FeedbackForm     FeedbackForm
	Title            string `gorm:"type:varchar(255); not null"`
	FeedbackFormID   uint   `gorm:"not null"`
	ForUserProfile   userModels.UserProfile
	ForUserProfileID uint
	ByProfile        userModels.UserProfile
	ByUserProfileID  uint   `gorm:"not null"`
	Team             userModels.Team
	TeamID           uint   `gorm:"not null"`
	Status           int8   `gorm:"default:0; not null"` // TODO Add enum
	SubmittedAt      time.Time
	DurationStart    time.Time
	DurationEnd      time.Time
	ExpireAt         time.Time
}
