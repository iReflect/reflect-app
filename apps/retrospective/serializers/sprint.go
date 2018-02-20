package serializers

import (
	"time"

	userSerializer "github.com/iReflect/reflect-app/apps/user/serializers"
)

// Sprint ...
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
}

// SprintsSerializer ...
type SprintsSerializer struct {
	Sprints []Sprint
}
