package serializers

import (
	"time"

	"github.com/iReflect/reflect-app/apps/feedback/models"
)

// FeedbackListSerializer lists the feedbacks for a given user
type FeedbackListSerializer struct {
	NewFeedbackCount       uint
	DraftFeedbackCount     uint
	SubmittedFeedbackCount uint
	Feedbacks              []models.Feedback
}

// FeedbackDetailSerializer returns the details of a feedback
type FeedbackDetailSerializer struct {
	ID             uint
	Title          string
	DurationStart  time.Time
	DurationEnd    time.Time
	SubmittedAt    time.Time
	ExpireAt       time.Time
	Status         int8
	FeedbackFormID uint
	Categories     map[uint]CategoryDetailSerializer
}

// QuestionResponseSerializer returns the question response
type QuestionResponseSerializer struct {
	Response string `json:"response"`
	Comment  string `json:"comment"`
}

// FeedbackResponseSerializer returns the feedback response
type FeedbackResponseSerializer struct {
	// Data is a 3-level nested structure (category -> skill -> question)
	Data map[uint]map[uint]map[uint]QuestionResponseSerializer `json:"data" binding:"required"`
	// SaveAndSubmit default value is false, i.e., if not present then it will be assumed false
	SaveAndSubmit bool `json:"saveAndSubmit"`
	Status        int8 `json:"status" binding:"required"`
}
