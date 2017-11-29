package models

import (
	"github.com/jinzhu/gorm"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

type TeamFeedbackForm struct {
	gorm.Model
	Team           userModels.Team
	TeamID         uint `gorm:"not null"`
	ForRole        userModels.Role
	ForRoleID      uint `gorm:"not null"`
	FeedbackForm   FeedbackForm
	FeedbackFormID uint `gorm:"not null"`
	Active         bool `gorm:"default:true; not null"`
}
