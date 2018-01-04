package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/iReflect/reflect-app/apps/feedback/models"
)

// Feedback represent a submitted/in-progress feedback form by a user
type Feedback struct {
	gorm.Model
	FeedbackForm     FeedbackForm
	Title            string `gorm:"type:varchar(255); not null"`
	FeedbackFormID   uint   `gorm:"not null"`
	ForUserProfile   UserProfile
	ForUserProfileID uint
	ByUserProfile    UserProfile
	ByUserProfileID  uint `gorm:"not null"`
	Team             Team
	TeamID           uint `gorm:"not null"`
	Status           models.QuestionType `gorm:"default:0; not null;"`
	SubmittedAt      *time.Time
	DurationStart    time.Time `gorm:"not null"`
	DurationEnd      time.Time `gorm:"not null"`
	ExpireAt         time.Time `gorm:"not null"`
}
