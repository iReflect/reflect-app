package models

import (
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/jinzhu/gorm"

	"time"
)

// RetrospectiveFeedbackScopeValues ...
var RetrospectiveFeedbackScopeValues = [...]string{
	"Team",
	"Individual",
}

// RetrospectiveFeedbackScope ...
type RetrospectiveFeedbackScope int8

// GetStringValue ...
func (status RetrospectiveFeedbackScope) GetStringValue() string {
	return RetrospectiveFeedbackScopeValues[status]
}

// RetrospectiveFeedbackScope
const (
	TeamScope RetrospectiveFeedbackScope = iota
	IndividualScope
)

// RetrospectiveFeedbackTypeValues ...
var RetrospectiveFeedbackTypeValues = [...]string{
	"Note",
	"Highlight",
	"Goal",
}

// RetrospectiveFeedbackType ...
type RetrospectiveFeedbackType int8

// GetStringValue ...
func (status RetrospectiveFeedbackType) GetStringValue() string {
	return RetrospectiveFeedbackTypeValues[status]
}

// RetrospectiveFeedbackType
const (
	NoteType RetrospectiveFeedbackType = iota
	HighlightType
	GoalType
)

// RetrospectiveFeedback represent Goals, Highlights and Notes of a sprint
type RetrospectiveFeedback struct {
	gorm.Model
	SubType         string                    `gorm:"type:varchar(30); not null"`
	Type            RetrospectiveFeedbackType `gorm:"default:0; not null"`
	Retrospective   Retrospective
	RetrospectiveID uint                       `gorm:"not null"`
	Text            string                     `gorm:"type:text; not null"`
	Scope           RetrospectiveFeedbackScope `gorm:"default:0; not null"`
	AssigneeID      *uint
	Assignee        *userModels.User
	AddedAt         *time.Time
	ResolvedAt      *time.Time
	ExpectedAt      *time.Time
	CreatedByID     uint `gorm:"not null"`
	CreatedBy       userModels.User
}

// TODO add admin support
