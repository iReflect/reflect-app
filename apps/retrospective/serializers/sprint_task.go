package serializers

import "time"

// SprintTask ...
type SprintTask struct {
	ID                   uint
	Key                  string
	URL                  string
	Summary              string
	Description          string
	Type                 string
	Status               string
	Priority             string
	IsTrackerTask        bool
	IsInvalid            bool
	Rating               int8
	Estimate             float64
	TotalPointsEarned    float64
	PointsEarned         float64
	Assignee             string // Assignee of the task
	Owner                string // Owner of the task across the sprints
	TaskParticipants     string // Participants of the task across the sprints
	SprintOwner          string // Owner of the task in the sprint
	SprintParticipants   string // Participants of the task in the current sprint
	SprintOwnerTime      uint   // Time spent on the task by the Sprint Owner in the sprint
	SprintOwnerTotalTime uint   // Total time spent on the task by the Sprint Owner across the sprints
	SprintTime           uint   // Time spent on the task in the sprint
	TotalTime            uint   // Total time spent on the task across the sprints
	DoneAt               *time.Time
	Resolution           int8
}

// SprintTasksSerializer ...
type SprintTasksSerializer struct {
	Tasks []*SprintTask
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
	Rating       int8
	Comment      string
	Role         int8
	Current      bool
}

// TaskMembersSerializer ...
type TaskMembersSerializer struct {
	Members []TaskMember
}

// SprintTaskUpdate ...
type SprintTaskUpdate struct {
	BaseRating
}

// SprintTaskDone ...
type SprintTaskDone struct {
	BaseResolution
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
	Role         *int8    `json:"Role" binding:"omitempty,is_valid_task_role"`
}
