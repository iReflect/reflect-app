package models

import (
	"github.com/jinzhu/gorm"
)

// TeamFeedbackForm represent the feedback template to be used for a role under a team
type TeamFeedbackForm struct {
	gorm.Model
	Team           Team
	TeamID         uint `gorm:"not null"`
	ForRole        Role
	ForRoleID      uint `gorm:"not null"`
	FeedbackForm   FeedbackForm
	FeedbackFormID uint `gorm:"not null"`
	Active         bool `gorm:"default:true; not null"`
}
