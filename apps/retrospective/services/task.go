package services

import (
	"errors"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/retrospective"
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
			"SUM(sprint_member_tasks.points_earned) over (PARTITION BY sm.member_id) as total_points, " +
			"SUM(sprint_member_tasks.points_earned) over (PARTITION BY sm.member_id, sm.sprint_id) as sprint_points, " +
			"SUM(sprint_member_tasks.time_spent_minutes) over (PARTITION BY sm.member_id) as total_time, " +
			"SUM(sprint_member_tasks.time_spent_minutes) over (PARTITION BY sm.member_id, sm.sprint_id) as sprint_time").
		QueryExpr()

	err = db.Raw("SELECT DISTINCT(smt.*) FROM (?) as smt WHERE smt.sprint_id = ?", dbs, sprintID).
		Scan(&members.Members).Error

	if err != nil {
		return nil, err
	}

	return members, nil
}

// GetMember returns the task member summary of a task for a particular sprint member
func (service TaskService) GetMember(sprintMemberTask retroModels.SprintMemberTask, memberID uint) (member *retroSerializers.TaskMember, err error) {
	db := service.DB
	member = new(retroSerializers.TaskMember)

	tempDB := db.Model(retroModels.SprintMemberTask{}).
		Where("task_id = ?", sprintMemberTask.TaskID).
		Joins("JOIN sprint_members AS sm ON sprint_member_tasks.sprint_member_id = sm.id AND sm.member_id = ?", memberID).
		Joins("JOIN users ON sm.member_id = users.id").
		Select("sprint_member_tasks.*," +
			"users.*," +
			"sm.sprint_id, " +
			"SUM(sprint_member_tasks.points_earned) over (PARTITION BY sprint_member_tasks.task_id) as total_points, " +
			"SUM(sprint_member_tasks.points_earned) over (PARTITION BY sprint_member_tasks.task_id, sm.sprint_id) as sprint_points, " +
			"SUM(sprint_member_tasks.time_spent_minutes) over (PARTITION BY sprint_member_tasks.task_id) as total_time, " +
			"SUM(sprint_member_tasks.time_spent_minutes) over (PARTITION BY sprint_member_tasks.task_id, sm.sprint_id) as sprint_time").
		QueryExpr()

	if err = db.Raw("SELECT DISTINCT(smt.*) FROM (?) as smt WHERE smt.sprint_member_id = ?", tempDB, sprintMemberTask.SprintMemberID).
		Scan(&member).Error; err != nil {
		return nil, err
	}

	return member, nil
}

// AddMember ...
func (service TaskService) AddMember(taskID string, retroID string, sprintID string, memberID uint) (member *retroSerializers.TaskMember, err error) {
	db := service.DB

	var sprintMember retroModels.SprintMember
	err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprintID).
		Where("member_id = ?", memberID).
		Find(&sprintMember).Error

	if err != nil {
		return nil, errors.New("member is not a part of the sprint")
	}

	err = db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_id = ?", sprintMember.ID).
		Where("task_id = ?", taskID).
		Find(&retroModels.SprintMemberTask{}).
		Error

	if err == nil {
		return nil, errors.New("member is already a part of the sprint task")
	}

	intTaskID, err := strconv.Atoi(taskID)
	if err != nil {
		return nil, err
	}

	sprintMemberTask := retroModels.SprintMemberTask{}
	sprintMemberTask.SprintMemberID = sprintMember.ID
	sprintMemberTask.TaskID = uint(intTaskID)
	sprintMemberTask.TimeSpentMinutes = 0
	sprintMemberTask.PointsEarned = 0
	sprintMemberTask.PointsAssigned = 0
	sprintMemberTask.Rating = retrospective.OkayRating
	sprintMemberTask.Comment = ""

	err = db.Create(&sprintMemberTask).Error
	if err != nil {
		return nil, err
	}

	member, err = service.GetMember(sprintMemberTask, memberID)
	return member, err
}

// UpdateTaskMember ...
func (service TaskService) UpdateTaskMember(taskID string, retroID string, sprintID string, taskMemberData *retroSerializers.SprintTaskMemberUpdate) (*retroSerializers.TaskMember, error) {
	db := service.DB

	sprintMemberTask := retroModels.SprintMemberTask{}
	err := db.Model(&retroModels.SprintMemberTask{}).
		Where("task_id = ?", taskID).
		Where("id = ?", taskMemberData.ID).
		Preload("SprintMember").
		Find(&sprintMemberTask).Error

	if err != nil {
		return nil, err
	}

	if taskMemberData.SprintPoints != nil {
		sprintMemberTask.PointsEarned = *taskMemberData.SprintPoints
	}
	if taskMemberData.Rating != nil {
		sprintMemberTask.Rating = retrospective.Rating(*taskMemberData.Rating)
	}
	if taskMemberData.Comment != nil {
		sprintMemberTask.Comment = *taskMemberData.Comment
	}
	if err = db.Save(&sprintMemberTask).Error; err != nil {
		return nil, err
	}
	return service.GetMember(sprintMemberTask, sprintMemberTask.SprintMember.MemberID)
}
