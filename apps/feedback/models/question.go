package models

import (
	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/db/models/fields"
)

type Question struct {
	gorm.Model
	Text    string       `gorm:"type:text; not null"`
	Type    int          `gorm:"default:0; not null"`
	Skill   Skill        `gorm:"ForeignKey:SkillID; AssociationForeignKey:ID"`
	SkillID uint         `gorm:"not null"`
	Options fields.JSONB `gorm:"type:jsonb; not null; default:'{}'::jsonb"`
	Weight  int          `gorm:"default:1; not null"`
}
