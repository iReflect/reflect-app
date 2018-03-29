package services

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/retrospective"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
	"net/http"
)

// TaskService ...
type TaskService struct {
	DB *gorm.DB
}

// List ...
func (service TaskService) List(
	retroID string,
	sprintID string) (taskList *retroSerializers.TasksSerializer, status int, err error) {
	db := service.DB
	taskList = new(retroSerializers.TasksSerializer)

	dbs := service.tasksForActiveAndCurrentSprint(retroID, sprintID).
		Select(`
            tasks.*,
            sm.sprint_id,
            SUM(smt.time_spent_minutes) OVER (PARTITION BY tasks.id)               AS total_time,
            SUM(smt.time_spent_minutes) OVER (PARTITION BY tasks.id, sm.sprint_id) AS sprint_time,
            SUM(smt.points_earned) OVER (PARTITION BY tasks.id)                    AS total_points_earned, 
            SUM(smt.points_earned) OVER (PARTITION BY tasks.id, sm.sprint_id )     AS points_earned
		`).
		QueryExpr()

	query := `
		SELECT 
			DISTINCT(t.*),
			CASE WHEN (t.total_points_earned > t.estimate + 0.05) THEN TRUE ELSE FALSE END AS is_invalid
		FROM (?) AS t WHERE t.sprint_id = ?
	`
	err = db.Raw(query, dbs, sprintID).Order("t.key").Scan(&taskList.Tasks).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get tasks")
	}

	return taskList, http.StatusOK, nil
}

// Get ...
func (service TaskService) Get(
	taskID string,
	retroID string,
	sprintID string) (task *retroSerializers.Task, status int, err error) {
	db := service.DB
	var tasks []retroSerializers.Task

	dbs := service.tasksForActiveAndCurrentSprint(retroID, sprintID).
		Where("tasks.id = ?", taskID).
		Select(`
            tasks.*,
            sm.sprint_id,
            SUM(smt.time_spent_minutes) OVER (PARTITION BY tasks.id) AS total_time,
            SUM(smt.time_spent_minutes) OVER (PARTITION BY tasks.id, sm.sprint_id) AS sprint_time`).
		QueryExpr()

	err = db.Raw("SELECT DISTINCT(t.*) FROM (?) AS t WHERE t.sprint_id = ?", dbs, sprintID).
		Scan(&tasks).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get task")
	}

	return &tasks[0], http.StatusOK, nil
}

// MarkDone ...
func (service TaskService) MarkDone(
	taskID string,
	retroID string,
	sprintID string) (task *retroSerializers.Task, status int, err error) {
	db := service.DB
	var sprint retroModels.Sprint
	err = db.Model(&retroModels.Sprint{}).
		Where("id = ?", sprintID).
		Scan(&sprint).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to mark the task as done")
	}

	err = db.Model(&retroModels.Task{}).
		Where("tasks.id = ?", taskID).
		Where("done_at is NULL").
		Update("done_at", *sprint.EndDate).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("task not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to mark the task as done")
	}

	return service.Get(taskID, retroID, sprintID)
}

// MarkUndone ...
func (service TaskService) MarkUndone(
	taskID string,
	retroID string,
	sprintID string) (task *retroSerializers.Task, status int, err error) {
	db := service.DB
	err = db.Model(&retroModels.Task{}).
		Where("tasks.id = ?", taskID).
		Where("done_at is not NULL").
		Update("done_at", nil).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("task not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to mark the task as done")
	}

	return service.Get(taskID, retroID, sprintID)
}

// GetMembers ...
func (service TaskService) GetMembers(
	taskID string,
	retroID string,
	sprintID string) (members *retroSerializers.TaskMembersSerializer, status int, err error) {
	db := service.DB
	members = new(retroSerializers.TaskMembersSerializer)

	dbs := service.smtForActiveAndCurrentSprint(taskID, sprintID).
		Select(`
            sprint_member_tasks.*,
            users.*,
            sm.sprint_id,
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sm.member_id) AS total_points,
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sm.member_id, sm.sprint_id) AS sprint_points,
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sm.member_id) AS total_time,
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sm.member_id, sm.sprint_id) AS sprint_time`).
		QueryExpr()

	err = db.Raw("SELECT DISTINCT(smt.*) FROM (?) AS smt WHERE smt.sprint_id = ?", dbs, sprintID).
		Order("smt.role, smt.first_name, smt.last_name").
		Scan(&members.Members).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get members")
	}

	return members, http.StatusOK, nil
}

// GetMember returns the task member summary of a task for a particular sprint member
func (service TaskService) GetMember(
	sprintMemberTask retroModels.SprintMemberTask,
	memberID uint, sprintID string) (member *retroSerializers.TaskMember, status int, err error) {
	db := service.DB
	member = new(retroSerializers.TaskMember)

	tempDB := service.smtForActiveAndCurrentSprint(fmt.Sprint(sprintMemberTask.TaskID), sprintID).
		Where("sm.member_id = ?", memberID).
		Select(`
            sprint_member_tasks.*,
            users.*, 
            sm.sprint_id, 
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_member_tasks.task_id) AS total_points, 
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_member_tasks.task_id, sm.sprint_id) AS sprint_points, 
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_member_tasks.task_id) AS total_time, 
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_member_tasks.task_id, sm.sprint_id) AS sprint_time`).
		QueryExpr()

	err = db.Raw("SELECT DISTINCT(smt.*) FROM (?) as smt WHERE smt.sprint_member_id = ?",
		tempDB,
		sprintMemberTask.SprintMemberID).
		Scan(&member).Error
	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get member")
	}

	return member, http.StatusOK, nil
}

// AddMember ...
func (service TaskService) AddMember(
	taskID string,
	retroID string,
	sprintID string,
	memberID uint) (member *retroSerializers.TaskMember, status int, err error) {
	db := service.DB

	var sprintMember retroModels.SprintMember
	err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprintID).
		Where("id = ?", memberID).
		Find(&sprintMember).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("member is not a part of the sprint")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get member summary")
	}

	err = db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_id = ?", sprintMember.ID).
		Where("task_id = ?", taskID).
		Find(&retroModels.SprintMemberTask{}).
		Error

	if err == nil {
		return nil, http.StatusBadRequest, errors.New("member is already a part of the sprint task")
	}

	intTaskID, err := strconv.Atoi(taskID)
	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusBadRequest, errors.New("invalid task id")
	}

	sprintMemberTask := retroModels.SprintMemberTask{}
	sprintMemberTask.SprintMemberID = sprintMember.ID
	sprintMemberTask.TaskID = uint(intTaskID)
	sprintMemberTask.TimeSpentMinutes = 0
	sprintMemberTask.PointsEarned = 0
	sprintMemberTask.PointsAssigned = 0
	sprintMemberTask.Rating = retrospective.DecentRating
	sprintMemberTask.Comment = ""

	err = db.Create(&sprintMemberTask).Error
	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get member summary")
	}

	return service.GetMember(sprintMemberTask, sprintMember.MemberID, sprintID)
}

// UpdateTaskMember ...
func (service TaskService) UpdateTaskMember(
	taskID string,
	retroID string,
	sprintID string,
	taskMemberData *retroSerializers.SprintTaskMemberUpdate) (*retroSerializers.TaskMember, int, error) {
	db := service.DB

	sprintMemberTask := retroModels.SprintMemberTask{}
	err := db.Model(&retroModels.SprintMemberTask{}).
		Where("task_id = ?", taskID).
		Where("id = ?", taskMemberData.ID).
		Preload("SprintMember").
		Find(&sprintMemberTask).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("task member not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to update task member")
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
	if taskMemberData.Role != nil {
		sprintMemberTask.Role = retroModels.MemberTaskRole(*taskMemberData.Role)
	}
	if err = db.Save(&sprintMemberTask).Error; err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to update task member")
	}
	return service.GetMember(sprintMemberTask, sprintMemberTask.SprintMember.MemberID, sprintID)
}

// tasksForActiveAndCurrentSprint ...
func (service TaskService) tasksForActiveAndCurrentSprint(retroID string, sprintID string) *gorm.DB {
	db := service.DB

	return db.Model(retroModels.Task{}).
		Where("tasks.retrospective_id = ?", retroID).
		Joins("JOIN sprint_member_tasks AS smt ON smt.task_id = tasks.id").
		Joins("JOIN sprint_members AS sm ON smt.sprint_member_id = sm.id").
		Joins("JOIN sprints ON sm.sprint_id = sprints.id").
		Where("(sprints.status <> ? OR sprints.id = ?)", retroModels.DraftSprint, sprintID).
		Scopes(retroModels.NotDeletedSprint)
}

// smtForActiveAndCurrentSprint ...
func (service TaskService) smtForActiveAndCurrentSprint(taskID string, sprintID string) *gorm.DB {
	db := service.DB

	return db.Model(retroModels.SprintMemberTask{}).
		Where("task_id = ?", taskID).
		Joins("JOIN sprint_members AS sm ON sprint_member_tasks.sprint_member_id = sm.id").
		Joins("JOIN users ON sm.member_id = users.id").
		Joins("JOIN sprints ON sm.sprint_id = sprints.id").
		Where("(sprints.status <> ? OR sprints.id = ?)", retroModels.DraftSprint, sprintID).
		Scopes(retroModels.NotDeletedSprint)
}
