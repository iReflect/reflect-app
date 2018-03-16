package models

import (
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/jinzhu/gorm"

	"time"
)

// SprintExtraScopeValues ...
var SprintExtraScopeValues = [...]string{
	"Team",
	"Individual",
}

// SprintExtraScope ...
type SprintExtraScope int8

// GetStringValue ...
func (status SprintExtraScope) GetStringValue() string {
	return SprintExtraScopeValues[status]
}

// SprintExtraScope
const (
	TeamScope SprintExtraScope = iota
	IndividualScope
)

// SprintExtraTypeValues ...
var SprintExtraTypeValues = [...]string{
	"Note",
	"Highlight",
	"Goal",
}

// SprintExtraType ...
type SprintExtraType int8

// GetStringValue ...
func (status SprintExtraType) GetStringValue() string {
	return SprintExtraTypeValues[status]
}

// SprintExtraType
const (
	NoteType SprintExtraType = iota
	HighlightType
	GoalType
)

// SprintExtra represent Goals, Highlights and Notes of a sprint
type SprintExtra struct {
	gorm.Model
	SubType         string          `gorm:"type:varchar(30); not null"`
	Type            SprintExtraType `gorm:"default:0; not null"`
	Retrospective   Retrospective
	RetrospectiveID uint             `gorm:"not null"`
	Text            string           `gorm:"type:text; not null"`
	Scope           SprintExtraScope `gorm:"default:0; not null"`
	AssigneeID      uint
	Assignee        userModels.User
	AddedAt         *time.Time
	ResolvedAt      *time.Time
	ExpectedAt      *time.Time
	CreatedByID     uint `gorm:"not null"`
	CreatedBy       userModels.User
}
