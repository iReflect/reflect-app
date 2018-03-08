package serializers

import (
	"time"

	userSerializer "github.com/iReflect/reflect-app/apps/user/serializers"
)

// Sprint is a serializer used in the Get sprint APIs
type Sprint struct {
	ID               uint
	Title            string
	SprintID         string
	Status           int8
	StartDate        time.Time
	EndDate          time.Time
	LastSyncedAt     *time.Time
	CurrentlySyncing bool
	CreatedBy        userSerializer.User
	CreatedByID      uint
	GoodHighlights   string
	OkayHighlights   string
	BadHighlights    string
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
	GoodHighlights *string
	OkayHighlights *string
	BadHighlights  *string
}
