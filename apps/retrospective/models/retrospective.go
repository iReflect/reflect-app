package models

import (
	"errors"

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
	StoryPointPerWeek  float64 `gorm:"not null"`
	CreatedBy          userModels.User
	CreatedByID        uint `gorm:"not null"`
}

// Validate ...
func (retrospective *Retrospective) Validate(db *gorm.DB) (err error) {
	if retrospective.StoryPointPerWeek < 0 {
		err = errors.New("story points per week cannot be negative")
		return err
	}
	return
}

// BeforeSave ...
func (retrospective *Retrospective) BeforeSave(db *gorm.DB) (err error) {
	return retrospective.Validate(db)
}

// BeforeUpdate ...
func (retrospective *Retrospective) BeforeUpdate(db *gorm.DB) (err error) {
	return retrospective.Validate(db)
}

// RegisterRetrospectiveToAdmin ...
func RegisterRetrospectiveToAdmin(Admin *admin.Admin, config admin.Config) {
	retrospective := Admin.AddResource(&Retrospective{}, &config)
	taskProviderConfigMeta := getTaskProviderConfigMetaFieldMeta()
	createdByMeta := userModels.GetUserFieldMeta("CreatedBy")

	retrospective.Meta(&taskProviderConfigMeta)
	retrospective.Meta(&createdByMeta)

	retrospective.IndexAttrs("-Sprints")
	retrospective.NewAttrs("-Sprints")
	retrospective.EditAttrs("-Sprints")
	retrospective.ShowAttrs("-Sprints")
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

// RetroJoinSprints ...
func RetroJoinSprints(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprints ON retrospectives.id = sprints.retrospective_id").Where("sprints.deleted_at IS NULL")
}

// RetroJoinTasks ...
func RetroJoinTasks(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN tasks ON retrospectives.id = tasks.retrospective_id").Where("tasks.deleted_at IS NULL")
}

// RetroJoinUserTeams ...
func RetroJoinUserTeams(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN user_teams ON retrospectives.team_id = user_teams.team_id").Where("user_teams.deleted_at IS NULL")
}
