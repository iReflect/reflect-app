package models

import (
	"github.com/jinzhu/gorm"
)

// SprintMemberTask ...
type SprintMemberTask struct {
	gorm.Model
	SprintMember     SprintMember
	SprintMemberID   uint `gorm:"not null"`
	Task             Task
	TaskID           uint    `gorm:"not null"`
	TimeSpentMinutes uint    `gorm:"not null"`
	PointsEarned     float64 `gorm:"default:0; not null"`
	PointsAssigned   float64 `gorm:"default:0; not null"`
	Rating           int8    `gorm:"default:2; not null"`
	Comment          string  `gorm:"type:text"`
}
