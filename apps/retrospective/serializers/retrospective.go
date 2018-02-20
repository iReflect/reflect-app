package serializers

import (
	"time"

	userSerializer "github.com/iReflect/reflect-app/apps/user/serializers"
	"github.com/iReflect/reflect-app/db/models/fields"
)

// Retrospective ...
type Retrospective struct {
	ID                 uint
	Title              string
	Team               userSerializer.Team
	TeamID             uint
	UpdatedAt          time.Time
	CreatedBy          userSerializer.User
	CreatedByID        uint
	TaskProviderConfig fields.JSONB
	HrsPerStoryPoint   float64
}

// RetrospectiveListSerializer ...
type RetrospectiveListSerializer struct {
	Retrospectives []Retrospective
}
