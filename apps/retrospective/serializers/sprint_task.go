package serializers

import "time"

// SprintTask ...
type SprintTask struct {
	ID                uint
	Key               string
	Summary           string
	Type              string
	Status            string
	Priority          string
	Assignee          string
	Owner             string
	Estimate          float64
	TotalTime         uint
	SprintTime        uint
	TotalPointsEarned float64
	PointsEarned      float64
	DoneAt            *time.Time
	IsTrackerTask     bool
	IsInvalid         bool
}

// SprintTasksSerializer ...
type SprintTasksSerializer struct {
	Tasks []SprintTask
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
	Rating       *int8
	Comment      string
	Role         *int8
}

// TaskMembersSerializer ...
type TaskMembersSerializer struct {
	Members []TaskMember
}

// AddSprintTaskMemberSerializer ...
type AddSprintTaskMemberSerializer struct {
	MemberID uint `json:"memberID" binding:"required"`
}

// SprintTaskMemberUpdate serializer to update
type SprintTaskMemberUpdate struct {
	BaseRating
	SprintPoints *float64 `json:"SprintPoints"`
	Comment      *string  `json:"Comment"`
	Role         *int8    `json:"Role" binding:"is_valid_task_role"`
}
