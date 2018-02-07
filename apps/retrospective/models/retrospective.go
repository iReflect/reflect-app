package models

import (
	"github.com/jinzhu/gorm"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/db/models/fields"
)

// Retrospective represents a retrospective of a team
type Retrospective struct {
	gorm.Model
	Title              string       `gorm:"type:varchar(255); not null"`
	TaskProviderConfig fields.JSONB `gorm:"type:jsonb; not null; default:'[]'::jsonb"`
	Team               userModels.Team
	TeamID             uint `gorm:"not null"`
	Sprints            []Sprint
	HrsPerStoryPoint   float64 `gorm:"not null"`
	CreatedBy          userModels.User
	CreatedByID        uint `gorm:"not null"`
}
