package models

import (
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
)

// SprintTask
type SprintTask struct {
	gorm.Model
	Sprint   Sprint
	SprintID uint `gorm:"not null"`
	Task     Task
	TaskID   uint `gorm:"not null"`
}

// RegisterSprintMemberTaskToAdmin ...
func RegisterSprintTaskToAdmin(Admin *admin.Admin, config admin.Config) {
	sprintTask := Admin.AddResource(&SprintTask{}, &config)
	taskMeta := getTaskMeta()
	sprintTask.Meta(&taskMeta)
}

// getTaskMeta ...
func getTaskMeta() admin.Meta {
	return admin.Meta{
		Name: "Task",
		Type: "select_one",
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			db := context.GetDB()
			var taskList []Task
			db.Model(&Task{}).Scan(&taskList)

			for _, value := range taskList {
				results = append(results, []string{strconv.Itoa(int(value.ID)), value.Key})
			}
			return
		},
	}
}

// STJoinTask ...
func STJoinTask(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN tasks ON sprint_tasks.task_id = tasks.id").
		Where("tasks.deleted_at IS NULL")
}

// STJoinSMT ...
func STJoinSMT(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprint_member_tasks ON sprint_tasks.id = sprint_member_tasks.sprint_task_id").
		Where("sprint_member_tasks.deleted_at IS NULL")
}

// STJoinSprint ...
func STJoinSprint(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprints ON sprint_tasks.sprint_id = sprints.id").
		Where("sprints.deleted_at IS NULL")
}
