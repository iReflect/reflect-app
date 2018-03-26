package serializers

import "time"

// Task ...
type Task struct {
	ID         uint
	TaskID     string
	Summary    string
	Type       string
	Status     string
	Priority   string
	Assignee   string
	Estimate   float64
	TotalTime  uint
	SprintTime uint
	DoneAt     *time.Time
	IsInvalid  bool
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
	Rating       *int8
	Comment      string
	Role         *int8
}

// TaskMembersSerializer ...
type TaskMembersSerializer struct {
	Members []TaskMember
}

// AddTaskMemberSerializer ...
type AddTaskMemberSerializer struct {
	MemberID uint `json:"memberID" binding:"required"`
}

// SprintTaskMemberUpdate serializer to update
type SprintTaskMemberUpdate struct {
	ID           uint     `json:"ID" binding:"required"`
	SprintPoints *float64 `json:"SprintPoints"`
	Rating       *int8    `json:"Rating" binding:"is_valid_rating"`
	Comment      *string  `json:"Comment"`
	Role         *int8    `json:"Role" binding:"is_valid_task_role"`
}
