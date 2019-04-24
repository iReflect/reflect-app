package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"github.com/iReflect/reflect-app/apps/timetracker"
)

// Team represent a team/project comprising a set of user
type Team struct {
	gorm.Model
	Name             string `gorm:"type:varchar(64);not null"`
	Description      string `gorm:"type:text"`
	Active           bool   `gorm:"default:true; not null"`
	TimeProviderName string `gorm:"not null"`
	Users            []User
}

// RegisterTeamToAdmin ...
func RegisterTeamToAdmin(Admin *admin.Admin, config admin.Config) {
	team := Admin.AddResource(&Team{}, &config)
	providerNameMeta := getTimeProviderMeta()

	team.Meta(&providerNameMeta)
	team.IndexAttrs("-Users")
	team.NewAttrs("-Users")
	team.EditAttrs("-Users")
	team.ShowAttrs("-Users")
}

// BeforeSave ...
func (team *Team) BeforeSave(scope *gorm.Scope) error {
	if _, exists := timetracker.TimeProvidersDisplayNameMap[team.TimeProviderName]; !exists {
		return errors.New("Invalid time provider name")
	}
	return nil
}

// getTimeProviderMeta ...
func getTimeProviderMeta() admin.Meta {
	return admin.Meta{
		Name: "TimeProviderName",
		Type: "select_one",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			team := value.(*Team)
			return team.TimeProviderName
		},
		FormattedValuer: func(value interface{}, context *qor.Context) interface{} {
			team := value.(*Team)
			return timetracker.TimeProvidersDisplayNameMap[team.TimeProviderName]
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			team := resource.(*Team)
			value := metaValue.Value.([]string)[0]
			team.TimeProviderName = value
		},
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			for key, value := range timetracker.TimeProvidersDisplayNameMap {
				results = append(results, []string{key, value})
			}
			return
		},
	}
}
