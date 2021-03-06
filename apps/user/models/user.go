package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"github.com/iReflect/reflect-app/apps/timetracker"
	"github.com/iReflect/reflect-app/db/models/fields"
)

// TimeProviderConfig ...
type TimeProviderConfig struct {
	Data interface{} `json:"data"`
	Type string      `json:"type"`
}

// User represent the app user in system
type User struct {
	gorm.Model
	Email              string       `gorm:"type:varchar(255); not null; unique_index"`
	FirstName          string       `gorm:"type:varchar(30); not null"`
	LastName           string       `gorm:"type:varchar(150)"`
	Password           []byte       `gorm:"type:bytea"`
	Active             bool         `gorm:"default:true; not null"`
	TimeProviderConfig fields.JSONB `gorm:"type:jsonb; not null; default:'{}'::jsonb"`
	IsAdmin            bool         `gorm:"default:false; not null"`
	Teams              []Team
	Profiles           []UserProfile
}

// BeforeSave ...
func (user *User) BeforeSave() error {
	return user.CleanUserData()
}

// CleanUserData ...
func (user *User) CleanUserData() error {
	var timeProviderConfigurations []TimeProviderConfig

	user.FirstName = strings.TrimSpace(user.FirstName)
	user.LastName = strings.TrimSpace(user.LastName)
	user.Email = strings.TrimSpace(user.Email)

	err := json.Unmarshal([]byte(user.TimeProviderConfig), &timeProviderConfigurations)
	if err != nil {
		return err
	}
	for index, timeConfig := range timeProviderConfigurations {
		timeProviderConfigurations[index].Type = strings.TrimSpace(timeConfig.Type)
		timeProviderConfigurations[index].Data = timetracker.CleanTimeProviderConfig(timeConfig.Data, timeProviderConfigurations[index].Type)
	}
	user.TimeProviderConfig, err = json.Marshal(timeProviderConfigurations)
	return err
}

// Stringify ...
func (user User) Stringify() string {
	return fmt.Sprintf("%v %v", user.FirstName, user.LastName)
}

// DisplayName ...
func (user User) DisplayName() string {
	return user.FirstName + " " + user.LastName
}

// RegisterUserToAdmin ...
func RegisterUserToAdmin(Admin *admin.Admin, config admin.Config) {
	user := Admin.AddResource(&User{}, &config)
	timeProviderConfigMeta := getTimeProviderConfigMetaFieldMeta()
	user.Meta(&timeProviderConfigMeta)

	user.IndexAttrs("-Teams", "-Profiles", "-Password")
	user.NewAttrs("-Teams", "-Profiles", "-Password")
	user.EditAttrs("-Teams", "-Profiles", "-Password")
	user.ShowAttrs("-Teams", "-Profiles", "-Password")
}

// GetUserFieldMeta ...
func GetUserFieldMeta(fieldName string) admin.Meta {
	return admin.Meta{
		Name: fieldName,
		Type: "select_one",
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			db := context.GetDB()
			var members []User
			db.Model(&User{}).Scan(&members)

			for _, value := range members {
				results = append(results, []string{strconv.Itoa(int(value.ID)), value.FirstName + " " + value.LastName})
			}
			return
		},
	}
}

// getTimeConfigMetaFieldMeta ...
func getTimeProviderConfigMetaFieldMeta() admin.Meta {
	return admin.Meta{
		Name: "TimeProviderConfig",
		Type: "text",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			user := value.(*User)
			return string(user.TimeProviderConfig)
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			user := resource.(*User)
			value := metaValue.Value.([]string)[0]
			user.TimeProviderConfig = fields.JSONB(value)
		}}
}
