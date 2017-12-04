package base_models

import (
	"github.com/jinzhu/gorm"
)

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
