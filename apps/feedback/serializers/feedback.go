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
	Response string `json:"response" binding:"isvalidquestionresponse"`
	Comment  string `json:"comment"`
}

// FeedbackResponseData is the type of question response which is provided in the feedback form submit API
type FeedbackResponseData map[int64]map[int64]map[int64]QuestionResponseSerializer

// FeedbackResponseSerializer returns the feedback response
type FeedbackResponseSerializer struct {
	// Data is a 3-level nested structure (category -> skill -> question, we would have to 'dive' to the last level to apply validations)
	Data FeedbackResponseData `json:"data" binding:"required,allquestionpresent,dive,dive,dive"`
	// SaveAndSubmit default value is false, i.e., if not present then it will be assumed false
	SaveAndSubmit bool   `json:"saveAndSubmit" binding:"isvalidsaveandsubmit"`
	Status        int8   `json:"status" binding:"required"`
	SubmittedAt   string `json:"submittedAt" binding:"isvalidsubmittedat"`
	FeedbackID    string
}
