package services

import (
	"errors"
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/retrospective"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
)

// SprintTaskService ...
type SprintTaskService struct {
	DB *gorm.DB
}

// List ...
func (service SprintTaskService) List(
	retroID string,
	sprintID string) (taskList *retroSerializers.SprintTasksSerializer, status int, err error) {
	db := service.DB
	taskList = new(retroSerializers.SprintTasksSerializer)

	dbs := service.tasksWithTimeDetailsForCurrentAndPrevSprint(retroID, sprintID, nil).
		QueryExpr()

	query := `
		SELECT 
			DISTINCT(t.*),
			CASE WHEN (t.total_points_earned > t.estimate + 0.05) THEN TRUE ELSE FALSE END AS is_invalid
		FROM (?) AS t WHERE t.sprint_id = ?
	`
	err = db.Raw(query, dbs, sprintID).Order("t.tracker_unique_id").Scan(&taskList.Tasks).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get issues")
	}

	return taskList, http.StatusOK, nil
}

// Get ...
func (service SprintTaskService) Get(
	sprintTaskID string,
	retroID string,
	sprintID string) (*retroSerializers.SprintTask, int, error) {
	db := service.DB
	var task retroSerializers.SprintTask

	dbs := service.tasksWithTimeDetailsForCurrentAndPrevSprint(retroID, sprintID, &sprintTaskID).
		QueryExpr()

	query := `
		SELECT 
			DISTINCT(t.*),
			CASE WHEN (t.total_points_earned > t.estimate + 0.05) THEN TRUE ELSE FALSE END AS is_invalid
		FROM (?) AS t WHERE t.sprint_id = ? AND t.id = ? 
	`

	err := db.Raw(query, dbs, sprintID, sprintTaskID).
		Scan(&task).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get issue")
	}

	return &task, http.StatusOK, nil
}

// Update ...
func (service SprintTaskService) Update(sprintTaskID string, retroID string, sprintID string, data retroSerializers.SprintTaskUpdate) (*retroSerializers.SprintTask, int, error) {
	db := service.DB

	var task retroModels.Task
	err := db.Model(&retroModels.Task{}).
		Where("tasks.deleted_at IS NULL").
		Scopes(retroModels.TaskJoinST).
		Where("sprint_tasks.id = ?", sprintTaskID).
		Where("sprint_tasks.sprint_id = ?", sprintID).
		Find(&task).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint task not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint task")
	}

	if data.Rating != nil {
		task.Rating = retrospective.Rating(*data.Rating)
	}

	if err := db.Save(&task).Error; err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to update sprint task")
	}

	return service.Get(sprintTaskID, retroID, sprintID)
}

// MarkDone ...
func (service SprintTaskService) MarkDone(
	sprintTaskID string,
	retroID string,
	sprintID string) (task *retroSerializers.SprintTask, status int, err error) {
	db := service.DB
	var sprint retroModels.Sprint
	err = db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Where("id = ?", sprintID).
		Scan(&sprint).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to mark the issue as done")
	}

	query := db.Model(&retroModels.SprintTask{}).
		Where("sprint_tasks.deleted_at IS NULL").
		Where("id = ?", sprintTaskID).
		Select("task_id").
		QueryExpr()
	err = db.Model(&retroModels.Task{}).
		Where("tasks.deleted_at IS NULL").
		Where("id = (?)", query).
		Update("done_at", gorm.Expr("COALESCE(done_at, ?)", *sprint.EndDate)).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("issue not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to mark the issue as done")
	}

	return service.Get(sprintTaskID, retroID, sprintID)
}

// MarkUndone ...
func (service SprintTaskService) MarkUndone(
	sprintTaskID string,
	retroID string,
	sprintID string) (task *retroSerializers.SprintTask, status int, err error) {
	db := service.DB
	query := db.Model(&retroModels.SprintTask{}).
		Where("sprint_tasks.deleted_at IS NULL").
		Where("id = ?", sprintTaskID).
		Select("task_id").
		QueryExpr()
	err = db.Model(&retroModels.Task{}).
		Where("tasks.deleted_at IS NULL").
		Where("id = (?)", query).
		Update("done_at", nil).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("issue not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to mark the task as done")
	}

	return service.Get(sprintTaskID, retroID, sprintID)
}

// tasksForCurrentAndPrevSprint ...
func (service SprintTaskService) tasksForCurrentAndPrevSprint(retroID string, sprintID string) *gorm.DB {
	db := service.DB

	currentSprintFilter := db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Where("sprints.id = ?", sprintID).
		Scopes(retroModels.NotDeletedSprint).
		Select("end_date").QueryExpr()

	sprintFilter := db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Scopes(retroModels.NotDeletedSprint).
		Where("retrospective_id = ? AND start_date < (?)", retroID, currentSprintFilter).
		Select("id").QueryExpr()

	return db.Model(retroModels.Task{}).
		Where("tasks.deleted_at IS NULL").
		Where("tasks.retrospective_id = ?", retroID).
		Scopes(
			retroModels.TaskJoinST,
			retroModels.STLeftJoinSMT,
			retroModels.SMTLeftJoinSM,
			retroModels.SMLeftJoinMember).
		Where("sprint_tasks.sprint_id in (?)", sprintFilter)
}

// tasksWithTimeDetailsForCurrentAndPrevSprint ...
func (service SprintTaskService) tasksWithTimeDetailsForCurrentAndPrevSprint(retroID string, sprintID string, sprintTaskID *string) *gorm.DB {

	db := service.DB

	taskOwnerTable := service.tasksForCurrentAndPrevSprint(retroID, sprintID).
		Select(`
        DISTINCT ON (tasks.id) tasks.id AS task_id,
        users.first_name || ' ' || users.last_name          AS member_name,
        SUM(sprint_member_tasks.time_spent_minutes)
        OVER (
          PARTITION BY tasks.id, sprint_members.member_id) AS member_time`).
		Order("tasks.id").
		Order("member_time DESC").
		Order("sprint_members.member_id DESC")

	tempSprintTaskMemberTable := service.tasksForCurrentAndPrevSprint(retroID, sprintID).
		Select(`
        tasks.id as temp_task_id,
		users.first_name || ' ' || users.last_name			AS sprint_member_name,
		sprint_member_tasks.time_spent_minutes,
		sprint_members.sprint_id,
		sprint_members.member_id,
        MAX(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY tasks.id, sprint_members.sprint_id) AS max_sprint_task_member_time,
        SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY tasks.id, sprint_members.member_id) AS sprint_task_member_total_time`)

	sprintTaskOwnerTable := db.Raw(`
        SELECT DISTINCT ON (temp_sprint_task_members.temp_task_id) temp_sprint_task_members.temp_task_id as task_id, *
		FROM (?) as temp_sprint_task_members
		WHERE temp_sprint_task_members.time_spent_minutes = temp_sprint_task_members.max_sprint_task_member_time
		AND temp_sprint_task_members.sprint_id = ?
		ORDER BY temp_sprint_task_members.temp_task_id, temp_sprint_task_members.sprint_task_member_total_time DESC,
		temp_sprint_task_members.member_id DESC`,
		tempSprintTaskMemberTable.QueryExpr(), sprintID)

	// TODO Update to include non-timesheet sprint tasks too
	dbs := service.tasksForCurrentAndPrevSprint(retroID, sprintID).
		Select(`
            sprint_tasks.id,
            tasks.key,
            tasks.tracker_unique_id,
            tasks.summary,
            tasks.description,
            tasks.type,
            tasks.status,
            tasks.priority,
            tasks.assignee,
            task_owners.member_name AS owner,
            sprint_task_owners.sprint_member_name as sprint_owner,
            sprint_task_owners.sprint_task_member_total_time as sprint_owner_total_time,
            sprint_task_owners.max_sprint_task_member_time as sprint_owner_time,
            tasks.estimate,
            tasks.rating,
            tasks.done_at,
            tasks.is_tracker_task,
            sprint_tasks.sprint_id,
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY tasks.id)                           AS total_time,
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY tasks.id, sprint_members.sprint_id) AS sprint_time,
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY tasks.id)                                AS total_points_earned, 
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY tasks.id, sprint_members.sprint_id )     AS points_earned
		`)

	if sprintTaskID != nil {
		sprintTaskFilter := db.Model(&retroModels.SprintTask{}).
			Where("sprint_tasks.deleted_at IS NULL").
			Where("id = ?", sprintTaskID).
			Select("task_id").QueryExpr()
		taskOwnerTable = taskOwnerTable.Where("sprint_tasks.task_id = (?)", sprintTaskFilter)
		dbs = dbs.Where("sprint_tasks.task_id = (?)", sprintTaskFilter)
	}
	dbs = dbs.
		Joins("LEFT JOIN (?) AS task_owners ON task_owners.task_id = tasks.id", taskOwnerTable.QueryExpr()).
		Joins(`LEFT JOIN (?) AS sprint_task_owners ON sprint_task_owners.task_id = tasks.id`, sprintTaskOwnerTable.QueryExpr())
	return dbs
}
