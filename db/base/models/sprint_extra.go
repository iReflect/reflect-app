package models

import (
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/jinzhu/gorm"

	"time"
)

// SprintExtra represent Goals, Highlights and Notes of a sprint
type SprintExtra struct {
	gorm.Model
	SubType         string `gorm:"type:varchar(30); not null"`
	Type            int8   `gorm:"default:0; not null"`
	Retrospective   Retrospective
	RetrospectiveID uint   `gorm:"not null"`
	Text            string `gorm:"type:text; not null"`
	Scope           int8   `gorm:"default:0; not null"`
	AssigneeID      uint
	Assignee        userModels.User
	AddedAt         *time.Time
	ResolvedAt      *time.Time
	ExpectedAt      *time.Time
	CreatedByID     uint `gorm:"not null"`
	CreatedBy       userModels.User
}
