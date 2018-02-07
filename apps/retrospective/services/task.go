package services

import (
	"github.com/jinzhu/gorm"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
)

//TaskService ...
type TaskService struct {
	DB *gorm.DB
}

// List ...
func (service TaskService) List(retroID string, sprintID string) (taskList *retroSerializers.TasksSerializer, err error) {
	db := service.DB
	taskList = new(retroSerializers.TasksSerializer)

	dbs := db.Model(retroModels.Task{}).
		Where("retrospective_id = ?", retroID).
		Joins("JOIN sprint_member_tasks AS smt ON smt.task_id = tasks.id").
		Joins("JOIN sprint_members AS sm ON smt.sprint_member_id = sm.id").
		Select("tasks.*, sm.sprint_id, SUM(smt.time_spent_minutes) over (PARTITION BY tasks.id) as total_time, SUM(smt.time_spent_minutes) over (PARTITION BY tasks.id, sm.sprint_id) as sprint_time").
		QueryExpr()

	err = db.Raw("SELECT DISTINCT(t.*) FROM (?) as t WHERE t.sprint_id = ?", dbs, sprintID).
		Scan(&taskList.Tasks).Error

	if err != nil {
		return nil, err
	}

	return taskList, nil
}
