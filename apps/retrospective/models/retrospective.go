package models

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/db/models/fields"
)

// Retrospective represents a retrospective of a team
type Retrospective struct {
	gorm.Model
	Title              string       `gorm:"type:varchar(255); not null"`
	ProjectName        string       `gorm:"type:varchar(255); not null"`
	TaskProviderConfig fields.JSONB `gorm:"type:jsonb; not null; default:'[]'::jsonb"`
	Team               userModels.Team
	TeamID             uint `gorm:"not null"`
	Sprints            []Sprint
	HrsPerStoryPoint   float64 `gorm:"not null"`
	CreatedBy          userModels.User
	CreatedByID        uint `gorm:"not null"`
}

// RegisterRetrospectiveToAdmin ...
func RegisterRetrospectiveToAdmin(Admin *admin.Admin, config admin.Config) {
	retrospective := Admin.AddResource(&Retrospective{}, &config)
	taskProviderConfigMeta := getTaskProviderConfigMetaFieldMeta()
	retrospective.Meta(&taskProviderConfigMeta)

}

// getTaskConfigMetaFieldMeta ...
func getTaskProviderConfigMetaFieldMeta() admin.Meta {
	return admin.Meta{
		Name: "TaskProviderConfig",
		Type: "text",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			retrospective := value.(*Retrospective)
			return string(retrospective.TaskProviderConfig)
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			retrospective := resource.(*Retrospective)
			value := metaValue.Value.([]string)[0]
			retrospective.TaskProviderConfig = fields.JSONB(value)
		}}
}
