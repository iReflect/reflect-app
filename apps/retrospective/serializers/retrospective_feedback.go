package serializers

import (
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/apps/user/serializers"
	"time"
)

// RetrospectiveFeedback ...
type RetrospectiveFeedback struct {
	ID              uint
	SubType         string
	Type            models.RetrospectiveFeedbackType
	RetrospectiveID uint
	Text            string
	Scope           models.RetrospectiveFeedbackScope
	AssigneeID      *uint
	Assignee        *serializers.User
	AddedAt         *time.Time
	ResolvedAt      *time.Time
	ExpectedAt      *time.Time
	CreatedByID     uint
	CreatedBy       serializers.User
}

// RetrospectiveFeedbackUpdateSerializer ...
type RetrospectiveFeedbackUpdateSerializer struct {
	Text       *string    `json:"Text"`
	Scope      *int8      `json:"Scope" binding:"is_valid_retrospective_feedback_scope"`
	AssigneeID *uint      `json:"AssigneeID"`
	ExpectedAt *time.Time `json:"ExpectedAt"`
}

// RetrospectiveFeedbackCreateSerializer ...
type RetrospectiveFeedbackCreateSerializer struct {
	SubType string `json:"subType" binding:"required"`
}

// RetrospectiveFeedbackListSerializer ...
type RetrospectiveFeedbackListSerializer struct {
	Feedbacks []models.RetrospectiveFeedback
}
