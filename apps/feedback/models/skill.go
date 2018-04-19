package models

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
)

// Skill represent the skill comprised by category
// TODO Add support for versioning and making it non-editable
type Skill struct {
	gorm.Model
	Title        string `gorm:"type:varchar(255); not null"`
	DisplayTitle string `gorm:"type:varchar(255)"`
	Description  string `gorm:"type:text"`
	Weight       int    `gorm:"default:1"`
	Questions    []Question
}

// RegisterSkillToAdmin ...
func RegisterSkillToAdmin(Admin *admin.Admin, config admin.Config) {
	skill := Admin.AddResource(&Skill{}, &config)
	questionsMeta := skill.Meta(&admin.Meta{Name: "Questions"})
	SetQuestionRelatedFieldMeta(questionsMeta.Resource)
	skill.IndexAttrs("-Questions")
	skill.NewAttrs("-Questions")
	skill.EditAttrs("-Questions")
	skill.ShowAttrs("-Questions")
}
