package models

import (
	"github.com/jinzhu/gorm"
	"time"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

type Feedback struct {
	gorm.Model
	FeedbackForm   FeedbackForm    `gorm:"ForeignKey:FeedbackFormID; AssociationForeignKey:ID"`
	FeedbackFormID uint            `gorm:"not null"`
	ForUser        userModels.User `gorm:"ForeignKey:ForUserID; AssociationForeignKey:ID"`
	ForUserID      uint
	ByUser         userModels.User `gorm:"ForeignKey:ByUserID; AssociationForeignKey:ID"`
	ByUserID       uint            `gorm:"not null"`
	Team           userModels.Team `gorm:"ForeignKey:TeamID; AssociationForeignKey:ID"`
	TeamID         uint            `gorm:"not null"`
	Status         int8            `gorm:"default:0; not null"`
	SubmittedAt    time.Time
	DurationStart  time.Time
	DurationEnd    time.Time
	ExpireAt       time.Time
}
