package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Sprint ...
type Sprint struct {
	gorm.Model
	Title            string `gorm:"type:varchar(255); not null"`
	SprintID         string `gorm:"type:varchar(30); not null"`
	Retrospective    Retrospective
	RetrospectiveID  uint `gorm:"not null"`
	Status           int8 `gorm:"default:0; not null"`
	StartDate        *time.Time
	EndDate          *time.Time
	SprintMembers    []SprintMember
	LastSyncedAt     *time.Time
	CurrentlySyncing bool `gorm:"default:true;not null"`
	CreatedBy        User
	CreatedByID      uint `gorm:"not null"`
}
