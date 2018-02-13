package serializers

import (
	"time"

	userSerializer "github.com/iReflect/reflect-app/apps/user/serializers"
)

// Retrospective ...
type Retrospective struct {
	ID          uint
	Title       string
	Team        userSerializer.Team
	TeamID      uint
	UpdatedAt   time.Time
	CreatedBy   userSerializer.User
	CreatedByID uint
}

// RetrospectiveListSerializer ...
type RetrospectiveListSerializer struct {
	Retrospectives []Retrospective
}

// Task ...
type Task struct {
	ID         uint
	TaskID     string
	Summary    string
	Type       string
	Status     string
	Assignee   string
	Estimate   float64
	TotalTime  uint
	SprintTime uint
}

// TasksSerializer ...
type TasksSerializer struct {
	Tasks []Task
}

// TaskMember ...
type TaskMember struct {
	ID           uint
	FirstName    string
	LastName     string
	TotalTime    uint
	SprintTime   uint
	TotalPoints  float64
	SprintPoints float64
	Rating       uint
	Comment      string
}

// TaskMembersSerializer ...
type TaskMembersSerializer struct {
	Members []TaskMember
}
