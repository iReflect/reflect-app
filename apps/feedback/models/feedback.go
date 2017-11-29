package models

import (
	"github.com/jinzhu/gorm"
	"time"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

type Feedback struct {
	gorm.Model
	FeedbackForm   FeedbackForm
	FeedbackFormID uint `gorm:"not null"`
	ForUser        userModels.User
	ForUserID      uint
	ByUser         userModels.User
	ByUserID       uint `gorm:"not null"`
	Team           userModels.Team
	TeamID         uint `gorm:"not null"`
	Status         int8 `gorm:"default:0; not null"`
	SubmittedAt    time.Time
	DurationStart  time.Time
	DurationEnd    time.Time
	ExpireAt       time.Time
}
