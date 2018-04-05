package models

import (
	"github.com/jinzhu/gorm"
)

// SprintTask
type SprintTask struct {
	gorm.Model
	Sprint   Sprint
	SprintID uint `gorm:"not null"`
	Task     Task
	TaskID   uint `gorm:"not null"`
}
