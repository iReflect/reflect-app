package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"github.com/iReflect/reflect-app/db/models/fields"
)

// Task represents the tasks for retrospectives
type Task struct {
	gorm.Model
	TaskID            string `gorm:"type:varchar(30); not null"`
	Retrospective     Retrospective
	RetrospectiveID   uint         `gorm:"not null"`
	Summary           string       `gorm:"type:varchar(255); not null"`
	Type              string       `gorm:"type:varchar(30); not null"`
	Status            string       `gorm:"type:varchar(50); not null"`
	Priority          string       `gorm:"type:varchar(50); not null"`
	Assignee          string       `gorm:"type:varchar(100); not null"`
	Estimate          float64      `gorm:"not null; default: 0"`
	Fields            fields.JSONB `gorm:"type:jsonb; not null; default:'{}'::jsonb"`
	DoneAt            *time.Time
	SprintMemberTasks []SprintMemberTask
}

// Stringify ...
func (task Task) Stringify() string {
	return fmt.Sprintf("%v", task.TaskID)
}

// RegisterTaskToAdmin ...
func RegisterTaskToAdmin(Admin *admin.Admin, config admin.Config) {
	task := Admin.AddResource(&Task{}, &config)
	taskProviderConfigMeta := getFieldsMetaFieldMeta()
	task.Meta(&taskProviderConfigMeta)

	task.IndexAttrs("-SprintMemberTasks")
	task.NewAttrs("-SprintMemberTasks")
	task.EditAttrs("-SprintMemberTasks")
	task.ShowAttrs("-SprintMemberTasks")
}

// getFieldsMetaFieldMeta ...
func getFieldsMetaFieldMeta() admin.Meta {
	return admin.Meta{
		Name: "Fields",
		Type: "text",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			task := value.(*Task)
			return string(task.Fields)
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			task := resource.(*Task)
			value := metaValue.Value.([]string)[0]
			task.Fields = fields.JSONB(value)
		}}
}
