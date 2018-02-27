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
