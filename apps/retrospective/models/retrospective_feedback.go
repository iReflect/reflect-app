package models

import (
	"errors"
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
	Assignee        userModels.User
	AddedAt         *time.Time
	ResolvedAt      *time.Time
	ExpectedAt      *time.Time
	CreatedByID     uint `gorm:"not null"`
	CreatedBy       userModels.User
}

// BeforeSave ...
func (feedback *RetrospectiveFeedback) BeforeSave(db *gorm.DB) (err error) {
	if feedback.ExpectedAt != nil && feedback.ExpectedAt.Before(*feedback.AddedAt) {
		err = errors.New("expected_at can not be before added at")
		return err
	}
	if feedback.ResolvedAt != nil && feedback.ResolvedAt.Before(*feedback.AddedAt) {
		err = errors.New("resolved_at can not be before added at")
		return err
	}

	if feedback.AssigneeID != nil {
		var userIds []uint
		if err = db.Raw("SELECT user_teams.user_id FROM user_teams JOIN retrospectives "+
			"ON retrospectives.team_id = user_teams.team_id WHERE retrospectives.id = ? "+
			"and user_teams.user_id = ?", feedback.RetrospectiveID, feedback.AssigneeID).
			Scan(&userIds).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("cannot assign to requested user")
			}
			return err
		}

	}
	return
}

// TODO add admin support
