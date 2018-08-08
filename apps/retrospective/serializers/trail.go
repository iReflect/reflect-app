package serializers

import (
	"time"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

// Trail . . . .
type Trail struct {
	Action       string
	ActionItem   string
	ActionItemID uint
	ActionBy     userModels.User
	ActionByID   uint
	CreatedAt    time.Time
}

// TrailSerializer used in trail API
type TrailSerializer struct {
	Trails []Trail
}
