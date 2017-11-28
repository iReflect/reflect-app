package models

import (
	"github.com/jinzhu/gorm"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

type TeamFeedbackForm struct {
	gorm.Model
	Team           userModels.Team `gorm:"ForeignKey:TeamID; AssociationForeignKey:ID"`
	TeamID         uint            `gorm:"not null"`
	Role           userModels.Role `gorm:"ForeignKey:RoleID; AssociationForeignKey:ID"`
	RoleID         uint            `gorm:"not null"`
	FeedbackForm   FeedbackForm    `gorm:"ForeignKey:FeedbackFormID; AssociationForeignKey:ID"`
	FeedbackFormID uint            `gorm:"not null"`
	Active         bool            `gorm:"default:true; not null"`
}
