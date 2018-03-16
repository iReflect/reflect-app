package models

import (
	"github.com/jinzhu/gorm"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

// Trail represents an action on a retrospective item
type Trail struct {
	gorm.Model
	Action       string `gorm:"type:varchar(255); not null"`
	ActionItem   string `gorm:"type:varchar(255); not null"`
	ActionItemID uint   `gorm:"not null"`
	ActionBy     userModels.User
	ActionByID   uint `gorm:"not null"`
}
