package serializers

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

// AddTaskMemberSerializer ...
type AddTaskMemberSerializer struct {
	MemberID uint `json:"memberID" binding:"required"`
}

// SprintTaskMemberUpdate serializer to update
type SprintTaskMemberUpdate struct {
	ID           uint    `json:"ID" binding:"required"`
	SprintPoints float64 `json:"SprintPoints" binding:"required"`
	Rating       int8    `json:"Rating" binding:"required,is_valid_rating"`
	Comment      string  `json:"Comment" binding:"required"`
}
