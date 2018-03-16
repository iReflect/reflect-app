package models

import (
	"github.com/jinzhu/gorm"
)

// Trail represents an action on a retrospective item
type Trail struct {
	gorm.Model
	Action       string `gorm:"type:varchar(255); not null"`
	ActionItem   string `gorm:"type:varchar(255); not null"`
	ActionItemID uint   `gorm:"not null"`
	ActionBy     User
	ActionByID   uint `gorm:"not null"`
}
