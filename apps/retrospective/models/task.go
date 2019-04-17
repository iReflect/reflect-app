package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"github.com/iReflect/reflect-app/apps/retrospective"
	"github.com/iReflect/reflect-app/db/models/fields"
)

// Task represents the tasks for retrospectives
type Task struct {
	gorm.Model
	Key               string `gorm:"type:varchar(30); not null"`
	TrackerUniqueID   string `gorm:"type:varchar(255); not null"`
	Retrospective     Retrospective
	RetrospectiveID   uint                 `gorm:"not null"`
	Summary           string               `gorm:"type:text; not null"`
	Description       string               `gorm:"type:text; not null"`
	Type              string               `gorm:"type:varchar(30); not null"`
	Status            string               `gorm:"type:varchar(50); not null"`
	Priority          string               `gorm:"type:varchar(50); not null"`
	Assignee          string               `gorm:"type:varchar(100); not null"`
	Estimate          float64              `gorm:"not null; default: 0"`
	Fields            fields.JSONB         `gorm:"type:jsonb; not null; default:'{}'::jsonb"`
	Rating            retrospective.Rating `gorm:"default:2; not null"`
	DoneAt            *time.Time
	IsTrackerTask     bool `gorm:"not null;default: false"`
	SprintMemberTasks []SprintMemberTask
	Resolution        int8 `gorm:"default:1"`
}

// Stringify ...
func (task Task) Stringify() string {
	return fmt.Sprintf("%v", task.Key)
}

// Validate ...
func (task Task) Validate(db *gorm.DB) (err error) {
	if task.ID != 0 && task.TrackerUniqueID == "" {
		return errors.New("tracker_unique_id cannot be empty")
	}
	return nil
}

// BeforeSave ...
func (task *Task) BeforeSave(db *gorm.DB) (err error) {
	return task.Validate(db)
}

// BeforeUpdate ...
func (task *Task) BeforeUpdate(db *gorm.DB) (err error) {
	return task.Validate(db)
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

// TaskJoinTaskKeyMaps ...
func TaskJoinTaskKeyMaps(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN task_key_maps ON task_key_maps.task_id = tasks.id AND task_key_maps.deleted_at IS NULL")
}

// TaskJoinST ...
func TaskJoinST(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprint_tasks ON tasks.id = sprint_tasks.task_id AND sprint_tasks.deleted_at IS NULL")
}
