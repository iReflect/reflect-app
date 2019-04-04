package models

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"github.com/iReflect/reflect-app/apps/tasktracker"
	"github.com/iReflect/reflect-app/apps/timetracker"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/db/models/fields"
	"github.com/iReflect/reflect-app/libs/utils"
)

// Retrospective represents a retrospective of a team
type Retrospective struct {
	gorm.Model
	Title              string       `gorm:"type:varchar(255); not null"`
	ProjectName        string       `gorm:"type:varchar(255); not null"`
	TaskProviderConfig fields.JSONB `gorm:"type:jsonb; not null; default:'[]'::jsonb"`
	TimeProviderName   string       `gorm:"not null"`
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
	if _, exists := timetracker.TimeProvidersDisplayNameMap[retrospective.TimeProviderName]; !exists {
		return errors.New("Invalid time provider name")
	}
	return
}

// BeforeSave ...
func (retrospective *Retrospective) BeforeSave(db *gorm.DB) (err error) {
	err = retrospective.Validate(db)
	return err
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
	providerNameMeta := getTimeProviderMeta()

	retrospective.Meta(&providerNameMeta)
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

// getTimeProviderMeta ...
func getTimeProviderMeta() admin.Meta {
	return admin.Meta{
		Name: "TimeProviderName",
		Type: "select_one",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			retrospective := value.(*Retrospective)
			return retrospective.TimeProviderName
		},
		FormattedValuer: func(value interface{}, context *qor.Context) interface{} {
			retrospective := value.(*Retrospective)
			return timetracker.TimeProvidersDisplayNameMap[retrospective.TimeProviderName]
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			retrospective := resource.(*Retrospective)
			value := metaValue.Value.([]string)[0]
			retrospective.TimeProviderName = value
		},
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			for key, value := range timetracker.TimeProvidersDisplayNameMap {
				results = append(results, []string{key, value})
			}
			return
		},
	}
}

// RetroJoinSprints ...
func RetroJoinSprints(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprints ON retrospectives.id = sprints.retrospective_id AND sprints.deleted_at IS NULL")
}

// RetroJoinTasks ...
func RetroJoinTasks(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN tasks ON retrospectives.id = tasks.retrospective_id AND tasks.deleted_at IS NULL")
}

// RetroJoinUserTeams ...
func RetroJoinUserTeams(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN user_teams ON retrospectives.team_id = user_teams.team_id AND user_teams.deleted_at IS NULL")
}

// GetTaskTrackerConnectionFromRetro ...
func GetTaskTrackerConnectionFromRetro(db *gorm.DB, retroID string) (tasktracker.Connection, error) {
	var retro Retrospective
	if err := db.Model(Retrospective{}).
		Where("retrospectives.deleted_at IS NULL").
		Where("id = ?", retroID).
		Find(&retro).Error; err != nil {
		utils.LogToSentry(err)
		return nil, fmt.Errorf("retrospective with ID %v not found", retroID)
	}

	taskProviderConfig, err := tasktracker.DecryptTaskProviders(retro.TaskProviderConfig)
	if err != nil {
		utils.LogToSentry(err)
		return nil, err
	}

	connection := tasktracker.GetConnection(taskProviderConfig)
	if connection == nil {
		err = errors.New("no valid connection found")
		utils.LogToSentry(err)
		return nil, err
	}
	return connection, nil
}
