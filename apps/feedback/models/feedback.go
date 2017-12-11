package models

import (
	"time"

	"github.com/jinzhu/gorm"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/db/models/fields"
)

// Feedback represent a submitted/in-progress feedback form by a user
type Feedback struct {
	gorm.Model
	FeedbackForm     FeedbackForm
	Title            string `gorm:"type:varchar(255); not null"`
	FeedbackFormID   uint   `gorm:"not null"`
	ForUserProfile   userModels.UserProfile
	ForUserProfileID uint
	ByUserProfile    userModels.UserProfile
	ByUserProfileID  uint `gorm:"not null"`
	Team             userModels.Team
	TeamID           uint `gorm:"not null"`
	Status           int8 `gorm:"default:0; not null"` // TODO Add enum
	SubmittedAt      *time.Time
	DurationStart    time.Time `gorm:"not null"`
	DurationEnd      time.Time `gorm:"not null"`
	ExpireAt         time.Time `gorm:"not null"`
}

// FeedbackListResponse lists the feedbacks for a given user
type FeedbackListResponse struct {
	NewFeedbackCount       uint
	DraftFeedbackCount     uint
	SubmittedFeedbackCount uint
	Feedbacks              []Feedback
}

// QuestionResponseDetail returns the question response for a particular question
type QuestionResponseDetail struct {
	ID         uint
	Text       string
	Type       int8
	Options    fields.JSONB
	Weight     int
	ResponseID uint
	Response   string
	Comment    string
}

// SkillQuestionList returns the skill details with a list of questions under that skill
type SkillQuestionList struct {
	ID           uint
	Title        string
	DisplayTitle string
	Description  string
	Weight       int
	Questions    []QuestionResponseDetail
}

// CategorySkillQuestions returns the list of questions for a skill of a category
type CategorySkillQuestions struct {
	ID          uint
	Title       string
	Description string
	Skills      map[uint]SkillQuestionList
}

// FeedbackDetailResponse returns the details of a feedback
type FeedbackDetailResponse struct {
	ID             uint
	Title          string
	DurationStart  time.Time
	DurationEnd    time.Time
	SubmittedAt    time.Time
	ExpireAt       time.Time
	Status         int8
	FeedbackFormID uint
	Categories     map[uint]CategorySkillQuestions
}
