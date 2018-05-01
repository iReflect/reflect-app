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

// GetMembers ...
func (service SprintTaskService) GetMembers(
	sprintTaskID string,
	retroID string,
	sprintID string) (members *retroSerializers.TaskMembersSerializer, status int, err error) {
	db := service.DB
	members = new(retroSerializers.TaskMembersSerializer)

	dbs := service.smtForCurrentAndPrevSprint(sprintTaskID, retroID, sprintID).
		Select(`
            DISTINCT ON (users.id)
            sprint_member_tasks.*,
            users.*,
            sprints.end_date AS sprint_end_date,
            sprint_members.sprint_id,
            CASE WHEN (sprint_members.sprint_id = ?) THEN TRUE ELSE FALSE END AS editable,
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_members.member_id)                                AS total_points,
            CASE WHEN (sprint_members.sprint_id = ?)
              THEN
                SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_members.member_id, sprint_members.sprint_id)
              ELSE
                0
              END                                                                                                              AS sprint_points,
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_members.member_id)                           AS total_time,
            CASE WHEN (sprint_members.sprint_id = ?)
              THEN
                SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_members.member_id, sprint_members.sprint_id)ELSE 
                0
              END                                                                                                              AS sprint_time`,
			sprintID, sprintID, sprintID).
		Order("users.id DESC, sprints.end_date DESC").
		QueryExpr()

	err = db.Raw("SELECT smt.* FROM (?) AS smt", dbs).
		Order("smt.editable DESC, smt.role, smt.first_name, smt.last_name").
		Scan(&members.Members).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get members")
	}

	return members, http.StatusOK, nil
}

// GetMember returns the task member summary of a task for a particular sprint member
func (service SprintTaskService) GetMember(
	sprintMemberTask retroModels.SprintMemberTask,
	memberID uint,
	retroID string,
	sprintID string) (member *retroSerializers.TaskMember, status int, err error) {
	db := service.DB
	member = new(retroSerializers.TaskMember)

	tempDB := service.smtForCurrentAndPrevSprint(fmt.Sprint(sprintMemberTask.SprintTaskID), retroID, sprintID).
		Where("sprint_members.member_id = ?", memberID).
		Select(`
            sprint_member_tasks.*,
            users.*, 
            sprint_members.sprint_id, 
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_tasks.task_id)                                AS total_points, 
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_tasks.task_id, sprint_members.sprint_id)      AS sprint_points, 
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_tasks.task_id)                           AS total_time, 
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_tasks.task_id, sprint_members.sprint_id) AS sprint_time`).
		QueryExpr()

	err = db.Raw("SELECT DISTINCT(smt.*), TRUE as editable FROM (?) as smt WHERE smt.sprint_member_id = ?",
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
func (service SprintTaskService) AddMember(
	sprintTaskID string,
	retroID string,
	sprintID string,
	memberID uint) (member *retroSerializers.TaskMember, status int, err error) {
	db := service.DB

	var sprintMember retroModels.SprintMember
	err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.deleted_at IS NULL").
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
		Where("sprint_member_tasks.deleted_at IS NULL").
		Where("sprint_member_id = ?", sprintMember.ID).
		Where("sprint_task_id = ?", sprintTaskID).
		Find(&retroModels.SprintMemberTask{}).
		Error

	if err == nil {
		return nil, http.StatusBadRequest, errors.New("member is already a part of the sprint task")
	}

	intSprintTaskID, err := strconv.Atoi(sprintTaskID)
	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusBadRequest, errors.New("invalid task id")
	}

	sprintMemberTask := retroModels.SprintMemberTask{}
	sprintMemberTask.SprintMemberID = sprintMember.ID
	sprintMemberTask.SprintTaskID = uint(intSprintTaskID)
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

	return service.GetMember(sprintMemberTask, sprintMember.MemberID, retroID, sprintID)
}

// UpdateTaskMember ...
func (service SprintTaskService) UpdateTaskMember(
	sprintTaskID string,
	retroID string,
	sprintID string,
	smtID string,
	taskMemberData *retroSerializers.SprintTaskMemberUpdate) (*retroSerializers.TaskMember, int, error) {
	db := service.DB

	sprintMemberTask := retroModels.SprintMemberTask{}
	err := db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_tasks.deleted_at IS NULL").
		Where("sprint_task_id = ?", sprintTaskID).
		Where("id = ?", smtID).
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
	if err = db.Set("gorm:save_associations", false).Save(&sprintMemberTask).Error; err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to update task member")
	}
	return service.GetMember(sprintMemberTask, sprintMemberTask.SprintMember.MemberID, retroID, sprintID)
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
          PARTITION BY tasks.id, sprint_members.member_id ) AS member_time`).
		Order("tasks.id").
		Order("member_time DESC")

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
	dbs = dbs.Joins("LEFT JOIN (?) AS task_owners ON task_owners.task_id = tasks.id", taskOwnerTable.QueryExpr())

	return dbs
}

// smtForCurrentAndPrevSprint ...
func (service SprintTaskService) smtForCurrentAndPrevSprint(sprintTaskID string, retroID string, sprintID string) *gorm.DB {
	db := service.DB

	sprintTaskFilter := db.Model(&retroModels.SprintTask{}).
		Where("sprint_tasks.deleted_at IS NULL").
		Where("id = ?", sprintTaskID).
		Select("task_id").QueryExpr()

	currentSprintFilter := db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Where("id = ?", sprintID).
		Scopes(retroModels.NotDeletedSprint).
		Select("end_date").QueryExpr()

	sprintFilter := db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Where("retrospective_id = ? AND start_date < (?)", retroID, currentSprintFilter).
		Select("id").QueryExpr()

	return db.Model(retroModels.SprintMemberTask{}).
		Where("sprint_member_tasks.deleted_at IS NULL").
		Where("sprint_tasks.task_id = (?)", sprintTaskFilter).
		Scopes(
			retroModels.SMTJoinST,
			retroModels.STJoinTask,
			retroModels.SMTJoinSM,
			retroModels.SMJoinSprint,
			retroModels.SMJoinMember).
		Where("sprints.id in (?)", sprintFilter).
		Scopes(retroModels.NotDeletedSprint)
}
