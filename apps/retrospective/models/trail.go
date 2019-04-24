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

// TrailJoinSM ...
func TrailJoinSM(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprint_members ON trails.action_item_id = sprint_members.id")
}

// TrailJoinST ...
func TrailJoinST(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprint_tasks ON trails.action_item_id = sprint_tasks.id")
}

// TrailJoinSMT ...
func TrailJoinSMT(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprint_member_tasks ON trails.action_item_id = sprint_member_tasks.id")
}

// TrailJoinFeedback ...
func TrailJoinFeedback(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN retrospective_feedbacks ON trails.action_item_id = retrospective_feedbacks.id")
}
