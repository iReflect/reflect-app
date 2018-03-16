package models

import (
	"github.com/jinzhu/gorm"
)

// SprintSyncStatus stores the sync history of a sprint
type SprintSyncStatus struct {
	gorm.Model
	SprintID uint `gorm:"not null"`
	Sprint   Sprint
	Status   int8 `gorm:"default:0; not null"`
}
