package services

import (
	"github.com/jinzhu/gorm"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
)

// TaskService ...
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
		Select("tasks.*, " +
			"sm.sprint_id, " +
			"SUM(smt.time_spent_minutes) over (PARTITION BY tasks.id) as total_time, " +
			"SUM(smt.time_spent_minutes) over (PARTITION BY tasks.id, sm.sprint_id) as sprint_time").
		QueryExpr()

	err = db.Raw("SELECT DISTINCT(t.*) FROM (?) as t WHERE t.sprint_id = ?", dbs, sprintID).
		Scan(&taskList.Tasks).Error

	if err != nil {
		return nil, err
	}

	return taskList, nil
}

// Get ...
func (service TaskService) Get(id string, retroID string, sprintID string) (task *retroSerializers.Task, err error) {
	db := service.DB
	tasks := []retroSerializers.Task{}

	dbs := db.Model(retroModels.Task{}).
		Where("retrospective_id = ?", retroID).
		Where("tasks.id = ?", id).
		Joins("JOIN sprint_member_tasks AS smt ON smt.task_id = tasks.id").
		Joins("JOIN sprint_members AS sm ON smt.sprint_member_id = sm.id").
		Select("tasks.*, " +
			"sm.sprint_id, " +
			"SUM(smt.time_spent_minutes) over (PARTITION BY tasks.id) as total_time, " +
			"SUM(smt.time_spent_minutes) over (PARTITION BY tasks.id, sm.sprint_id) as sprint_time").
		QueryExpr()

	err = db.Raw("SELECT DISTINCT(t.*) FROM (?) as t WHERE t.sprint_id = ?", dbs, sprintID).
		Scan(&tasks).Error

	if err != nil {
		return nil, err
	}

	return &tasks[0], nil
}

// GetMembers ...
func (service TaskService) GetMembers(id string, retroID string, sprintID string) (members *retroSerializers.TaskMembersSerializer, err error) {
	db := service.DB
	members = new(retroSerializers.TaskMembersSerializer)

	dbs := db.Model(retroModels.SprintMemberTask{}).
		Where("task_id = ?", id).
		Joins("JOIN sprint_members AS sm ON sprint_member_tasks.sprint_member_id = sm.id").
		Joins("JOIN users ON sm.member_id = users.id").
		Select("sprint_member_tasks.*," +
			"users.*," +
			"sm.sprint_id, " +
			"SUM(sprint_member_tasks.points_earned) over (PARTITION BY sprint_member_tasks.sprint_member_id) as total_points, " +
			"SUM(sprint_member_tasks.points_earned) over (PARTITION BY sprint_member_tasks.sprint_member_id, sm.sprint_id) as sprint_points, " +
			"SUM(sprint_member_tasks.time_spent_minutes) over (PARTITION BY sprint_member_tasks.task_id) as total_time, " +
			"SUM(sprint_member_tasks.time_spent_minutes) over (PARTITION BY sprint_member_tasks.task_id, sm.sprint_id) as sprint_time").
		QueryExpr()

	err = db.Raw("SELECT DISTINCT(smt.*) FROM (?) as smt WHERE smt.sprint_id = ?", dbs, sprintID).
		Scan(&members.Members).Error

	if err != nil {
		return nil, err
	}

	return members, nil
}
