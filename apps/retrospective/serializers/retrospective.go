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
	ProjectName        string
	Team               userSerializer.Team
	TeamID             uint
	UpdatedAt          time.Time
	CreatedBy          userSerializer.User
	CreatedByID        uint
	CreatedAt          time.Time
	TaskProviderConfig fields.JSONB
	TimeProviderName   string
	StoryPointPerWeek  float64
}

// EditLevel ...
type EditLevel uint

// different EditLevels ...
const (
	NotEditable EditLevel = iota + 1
	Partially
	Fully
)

// RetroFieldsEditLevelMap ...
type RetroFieldsEditLevelMap struct {
	Title             EditLevel
	ProjectName       EditLevel
	TeamID            EditLevel
	StoryPointPerWeek EditLevel
	TimeProviderName  EditLevel
}

func (editLevel EditLevel) validate() bool {
	if editLevel == NotEditable || editLevel == Partially || editLevel == Fully {
		return true
	}
	return false
}

// Validate ...
func (retroFieldsEditLevelMap RetroFieldsEditLevelMap) Validate() bool {
	if retroFieldsEditLevelMap.ProjectName.validate() &&
		retroFieldsEditLevelMap.StoryPointPerWeek.validate() &&
		retroFieldsEditLevelMap.TeamID.validate() &&
		retroFieldsEditLevelMap.Title.validate() {
		return true
	}
	return false
}

// RetrospectiveCreateSerializer ...
type RetrospectiveCreateSerializer struct {
	Title              string                   `json:"title" binding:"required"`
	ProjectName        string                   `json:"projectName" binding:"required"`
	TaskProviderConfig []map[string]interface{} `json:"taskProvider" binding:"required,is_valid_task_provider_config"`
	TeamID             uint                     `json:"team" binding:"required,is_valid_team"`
	StoryPointPerWeek  float64                  `json:"storyPointPerWeek" binding:"required"`
	TimeProviderName   string                   `json:"timeProviderKey" binding:"required"`
	CreatedByID        uint
}

// RetrospectiveUpdateSerializer ...
type RetrospectiveUpdateSerializer struct {
	RetroID            uint                     `json:"retroID" binding:"required"`
	Title              string                   `json:"title" binding:"required"`
	ProjectName        string                   `json:"projectName" binding:"required"`
	TaskProviderConfig []map[string]interface{} `json:"taskProvider" binding:"required,is_valid_task_provider_config"`
	TeamID             uint                     `json:"team" binding:"required,is_valid_team"`
	StoryPointPerWeek  float64                  `json:"storyPointPerWeek" binding:"required"`
	TimeProviderName   string                   `json:"timeProviderKey" binding:"required"`
	CredentialsChanged bool                     `json:"credentialsChanged"`
}

// RetrospectiveListSerializer ...
type RetrospectiveListSerializer struct {
	Retrospectives []Retrospective
}
