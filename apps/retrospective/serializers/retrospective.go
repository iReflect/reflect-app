package serializers

import (
	"time"

	userSerializer "github.com/iReflect/reflect-app/apps/user/serializers"
)

// Retrospective ...
type Retrospective struct {
	ID        uint
	Title     string
	Team      userSerializer.Team
	TeamID    uint
	UpdatedAt time.Time
	CreatedBy userSerializer.User
	CreatedByID    uint
}

// RetroSpectiveListSerializer ...
type RetroSpectiveListSerializer struct {
	RetroSpectives []Retrospective
}
