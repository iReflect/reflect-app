package models

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"github.com/iReflect/reflect-app/db/models/fields"
)

// User represent the app user in system
type User struct {
	gorm.Model
	Name               string `gorm:"type:varchar(255); not null"`
	Email              string `gorm:"type:varchar(255); not null; unique_index"`
	FirstName          string `gorm:"type:varchar(30); not null"`
	LastName           string `gorm:"type:varchar(150)"`
	Active             bool   `gorm:"default:true; not null"`
	Teams              []Team
	Profiles           []UserProfile
	TimeProviderConfig fields.JSONB `gorm:"type:jsonb; not null; default:'{}'::jsonb"`
	IsAdmin            bool         `gorm:"default:false; not null"`
}

// RegisterUserToAdmin ...
func RegisterUserToAdmin(Admin *admin.Admin, config admin.Config) {
	user := Admin.AddResource(&User{}, &config)
	user.IndexAttrs("-Teams", "-Profiles")
	user.NewAttrs("-Teams", "-Profiles")
	user.EditAttrs("-Teams", "-Profiles")
	user.ShowAttrs("-Teams", "-Profiles")
	timeProviderConfigMeta := getTimeProviderConfigMetaFieldMeta()
	user.Meta(&timeProviderConfigMeta)
}

// AfterFind ...
func (u *User) AfterFind() (err error) {
	u.Name = u.FirstName + " " + u.LastName
	return nil
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
