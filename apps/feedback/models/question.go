package models

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"github.com/iReflect/reflect-app/db/models/fields"
)

// Question represent the questions asked for a skill
// TODO Add support for versioning and making it non-editable
type Question struct {
	gorm.Model
	Text    string `gorm:"type:text; not null"`
	Type    int8   `gorm:"default:0; not null"` // TODO Add enum
	Skill   Skill
	SkillID uint         `gorm:"not null"`
	Options fields.JSONB `gorm:"type:jsonb; not null; default:'{}'::jsonb"`
	Weight  int          `gorm:"default:1; not null"`
}

func RegisterQuestionToAdmin(Admin *admin.Admin, config admin.Config) {
	question := Admin.AddResource(&Question{}, &config)
	optionsMeta := getOptionsFieldMeta()

	question.Meta(&optionsMeta)

}

func SetQuestionRelatedFieldMeta(res *admin.Resource) {
	optionsMeta := getOptionsFieldMeta()
	res.Meta(&optionsMeta)
}

func getOptionsFieldMeta() admin.Meta {
	return admin.Meta{
		Name: "Options",
		Type: "text",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			question := value.(*Question)
			return string(question.Options)
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			question := resource.(*Question)
			value := metaValue.Value.([]string)[0]
			question.Options = fields.JSONB(value)
		}}

}
