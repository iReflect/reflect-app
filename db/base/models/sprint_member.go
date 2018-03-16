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
	MemberID           uint    `gorm:"not null"`
	AllocationPercent  float64 `gorm:"not null;default:100"`
	ExpectationPercent float64 `gorm:"not null;default:100"`
	Tasks              []SprintMemberTask
	Vacations          float64 `gorm:"not null;default:0"`
	Rating             int8    `gorm:"default:2; not null"`
	Comment            string  `gorm:"type:text"`
}
