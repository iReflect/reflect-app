package models

import (
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"github.com/iReflect/reflect-app/db/models/fields"
)

// User represent the app user in system
type User struct {
	gorm.Model
	Email              string `gorm:"type:varchar(255); not null; unique_index"`
	FirstName          string `gorm:"type:varchar(30); not null"`
	LastName           string `gorm:"type:varchar(150)"`
	Password           string `gorm:"type:varchar(255)"`
	Active             bool   `gorm:"default:true; not null"`
	Teams              []Team
	Profiles           []UserProfile
	TimeProviderConfig fields.JSONB `gorm:"type:jsonb; not null; default:'{}'::jsonb"`
	IsAdmin            bool         `gorm:"default:false; not null"`
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

	user.IndexAttrs("-Teams", "-Profiles")
	user.NewAttrs("-Teams", "-Profiles")
	user.EditAttrs("-Teams", "-Profiles")
	user.ShowAttrs("-Teams", "-Profiles")
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
