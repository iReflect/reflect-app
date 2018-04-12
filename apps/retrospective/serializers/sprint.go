package serializers

import (
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	"time"

	userSerializer "github.com/iReflect/reflect-app/apps/user/serializers"
)

// Sprint is a serializer used in the Get sprint APIs
type Sprint struct {
	ID              uint
	Title           string
	SprintID        string
	Status          retroModels.SprintStatus
	StartDate       time.Time
	EndDate         time.Time
	LastSyncedAt    *time.Time
	SyncStatus      int8
	CreatedBy       userSerializer.User
	CreatedByID     uint
	RetrospectiveID uint
	Summary         SprintSummary
	Editable        *bool
}

// SetEditable ...
func (sprint *Sprint) SetEditable(userID uint) {
	 *sprint.Editable = sprint.Status == retroModels.ActiveSprint ||
		(sprint.Status == retroModels.DraftSprint && sprint.CreatedByID == userID)
}

// SprintSummary ...
type SprintSummary struct {
	MemberCount      int
	TotalAllocation  float64
	TotalExpectation float64
	TotalOutput      float64
	Holidays         float64
	TotalVacations   float64
	TargetSP         float64
	TaskSummary      map[string]SprintTaskSummary
}

// SprintTaskSummary ...
type SprintTaskSummary struct {
	Count             uint
	TotalCount        uint
	PointsEarned      float64
	TotalPointsEarned float64
}

// SprintsSerializer ...
type SprintsSerializer struct {
	Sprints []Sprint
}

// AddMemberSerializer ...
type AddMemberSerializer struct {
	MemberID uint `json:"memberID" binding:"required"`
}

// CreateSprintSerializer is used in sprint create API
type CreateSprintSerializer struct {
	Title       string     `json:"title" binding:"required"`
	SprintID    string     `json:"sprintID" binding:"is_valid_sprint"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	CreatedByID uint
}

// UpdateSprintSerializer is used in sprint create API
type UpdateSprintSerializer struct {
}
