package models

import (
	"github.com/jinzhu/gorm"
)

// TaskKeyMap ...
type TaskKeyMap struct {
	gorm.Model
	TaskID uint `gorm:"not null"`
	Task   Task
	Key    string `gorm:"type:varchar(30); not null"`
}
