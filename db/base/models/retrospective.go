package models

import (
	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/db/models/fields"
)

//Retrospective ...
type Retrospective struct {
	gorm.Model
	Title              string       `gorm:"type:varchar(255); not null"`
	ProjectName        string       `gorm:"type:varchar(255); not null"`
	TaskProviderConfig fields.JSONB `gorm:"type:jsonb; not null; default:'[]'::jsonb"`
	Team               Team
	TeamID             uint `gorm:"not null"`
	Sprints            []Sprint
	HrsPerStoryPoint   float64 `gorm:"not null"`
	CreatedBy          User
	CreatedByID        uint `gorm:"not null"`
}
