package serializers

import (
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/apps/user/serializers"
	"time"
)

// RetrospectiveFeedbackSerializer ...
type RetrospectiveFeedbackSerializer struct {
	ID              uint
	SubType         string
	Type            models.RetrospectiveFeedbackType
	RetrospectiveID uint
	Text            string
	Scope           models.RetrospectiveFeedbackScope
	Assignee        *serializers.User
	AddedAt         *time.Time
	ResolvedAt      *time.Time
	ExpectedAt      *time.Time
	CreatedBy       serializers.User
}

// RetrospectiveFeedbackUpdateSerializer ...
type RetrospectiveFeedbackUpdateSerializer struct {
	Text       *string    `json:"Text"`
	Scope      *int8      `json:"Scope"`
	AssigneeID *uint      `json:"AssigneeID"`
	ExpectedAt *time.Time `json:"ExpectedAt"`
}

// RetrospectiveFeedbackCreateSerializer ...
type RetrospectiveFeedbackCreateSerializer struct {
	SubType string `json:"subType" binding:"required"`
}

// RetrospectiveFeedbackListSerializer ...
type RetrospectiveFeedbackListSerializer struct {
	Feedbacks []*RetrospectiveFeedbackSerializer
}