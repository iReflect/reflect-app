package serializers

import (
	"github.com/iReflect/reflect-app/apps/feedback/models"
	"time"
)

// FeedbackListResponse lists the feedbacks for a given user
type FeedbackListSerializer struct {
	NewFeedbackCount       uint
	DraftFeedbackCount     uint
	SubmittedFeedbackCount uint
	Feedbacks              []models.Feedback
	Token                  string
}

// FeedbackDetailResponse returns the details of a feedback
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
