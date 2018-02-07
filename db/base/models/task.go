package models

import (
	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/db/models/fields"
)

// Task ...
type Task struct {
	gorm.Model
	TaskID          string `gorm:"type:varchar(30); not null"`
	Retrospective   Retrospective
	RetrospectiveID uint         `gorm:"not null"`
	Summary         string       `gorm:"type:varchar(255); not null"`
	Type            string       `gorm:"type:varchar(30); not null"`
	Status          string       `gorm:"type:varchar(50); not null"`
	Assignee        string       `gorm:"type:varchar(100); not null"`
	Estimate        uint         `gorm:"not null"`
	Fields          fields.JSONB `gorm:"type:jsonb; not null; default:'{}'::jsonb"`
}
