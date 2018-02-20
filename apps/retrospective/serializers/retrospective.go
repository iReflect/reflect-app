package serializers

import (
	"time"

	userSerializer "github.com/iReflect/reflect-app/apps/user/serializers"
	"github.com/iReflect/reflect-app/db/models/fields"
)

// Retrospective ...
type Retrospective struct {
	ID                 uint
	Title              string
	Team               userSerializer.Team
	TeamID             uint
	UpdatedAt          time.Time
	CreatedBy          userSerializer.User
	CreatedByID        uint
	TaskProviderConfig fields.JSONB
	HrsPerStoryPoint   float64
}

// RetrospectiveCreateSerializer ...
type RetrospectiveCreateSerializer struct {
	Title              string                   `json:"title" binding:"required"`
	TaskProviderConfig []map[string]interface{} `json:"taskProvider" binding:"required,is_valid_task_provider_config"`
	TeamID             uint                     `json:"team" binding:"required,is_valid_team"`
	HrsPerStoryPoint   float64                  `json:"hoursPerStoryPoint" binding:"required"`
	CreatedByID        uint
}

// RetrospectiveListSerializer ...
type RetrospectiveListSerializer struct {
	Retrospectives []Retrospective
}
