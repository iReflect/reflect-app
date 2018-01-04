package models

import (
	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/db/models/fields"
	"github.com/iReflect/reflect-app/apps/feedback/models"
)

// Question represent the questions asked for a skill
// TODO Add support for versioning and making it non-editable
type Question struct {
	gorm.Model
	Text    string `gorm:"type:text; not null"`
	Type    models.QuestionType   `gorm:"default:0; not null"`
	Skill   Skill
	SkillID uint         `gorm:"not null"`
	Options fields.JSONB `gorm:"type:jsonb; not null; default:'{}'::jsonb"`
	Weight  int          `gorm:"default:1; not null"`
}
