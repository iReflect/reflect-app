package models

import (
	"github.com/jinzhu/gorm"
)

// SprintMember ...
type SprintMember struct {
	gorm.Model
	Sprint             Sprint
	SprintID           uint `gorm:"not null"`
	Member             User
	MemberID           uint `gorm:"not null"`
	AllocationPercent  uint `gorm:"not null"`
	ExpectationPercent uint `gorm:"not null"`
	Tasks              []SprintMemberTask
	Vacations          uint   `gorm:"not null;default:0"`
	Rating             int8   `gorm:"default:2; not null"`
	Comment            string `gorm:"type:text; not null"`
}
