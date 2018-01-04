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
	Status         models.FeedbackStatus
	FeedbackFormID uint
	Categories     map[uint]CategoryDetailSerializer
}

// FeedbackResponseData is the type of question response which is provided in the feedback form submit API
// It is a 3-level nested structure (category -> skill -> question),
// therefore we would have to 'dive' to the last level to apply any validations
type FeedbackResponseData map[int64]map[int64]map[int64]QuestionResponseSerializer

// FeedbackResponseSerializer returns the feedback response
type FeedbackResponseSerializer struct {
	Data        FeedbackResponseData  `json:"data" binding:"required,all_questions_present,dive,dive,dive"`
	Status      models.FeedbackStatus `json:"status" binding:"required"`
	SubmittedAt string                `json:"submittedAt" binding:"is_valid_submitted_at"`
	FeedbackID  string
}
