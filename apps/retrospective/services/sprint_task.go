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

	dbs := service.tasksForActiveAndCurrentSprint(retroID, sprintID, nil).
		Select(`
            sprint_tasks.id,
            tasks.key,       
            tasks.summary,   
            tasks.type,      
            tasks.status,    
            tasks.priority,  
            tasks.assignee,  
            tasks.estimate,  
            tasks.done_at,    
            sprint_members.sprint_id,
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY tasks.id)                           AS total_time,
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY tasks.id, sprint_members.sprint_id) AS sprint_time,
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY tasks.id)                                AS total_points_earned, 
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY tasks.id, sprint_members.sprint_id )     AS points_earned
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
func (service SprintTaskService) Get(
	sprintTaskID string,
	retroID string,
	sprintID string) (task *retroSerializers.SprintTask, status int, err error) {
	db := service.DB
	var tasks []retroSerializers.SprintTask

	dbs := service.tasksForActiveAndCurrentSprint(retroID, sprintID, &sprintTaskID).
		Where("tasks.id = ?", sprintTaskID).
		Select(`
            sprint_tasks.id,
            tasks.key,       
            tasks.summary,   
            tasks.type,      
            tasks.status,    
            tasks.priority,  
            tasks.assignee,  
            tasks.estimate,  
            tasks.done_at,    
            sprint_members.sprint_id,
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY tasks.id) AS total_time,
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY tasks.id, sprint_members.sprint_id) AS sprint_time`).
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
func (service SprintTaskService) MarkDone(
	sprintTaskID string,
	retroID string,
	sprintID string) (task *retroSerializers.SprintTask, status int, err error) {
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

	query := db.Model(&retroModels.SprintTask{}).Where("id = ?", sprintTaskID).Select("id").QueryExpr()
	err = db.Model(&retroModels.Task{}).
		Where("id = (?)", query).
		Update("done_at", gorm.Expr("COALESCE(done_at, ?)", *sprint.EndDate)).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("task not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to mark the task as done")
	}

	return service.Get(sprintTaskID, retroID, sprintID)
}

// MarkUndone ...
func (service SprintTaskService) MarkUndone(
	sprintTaskID string,
	retroID string,
	sprintID string) (task *retroSerializers.SprintTask, status int, err error) {
	db := service.DB
	query := db.Model(&retroModels.SprintTask{}).Where("id = ?", sprintTaskID).Select("id").QueryExpr()
	err = db.Model(&retroModels.Task{}).
		Where("id = (?)", query).
		Update("done_at", nil).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("task not found")
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

	dbs := service.smtForActiveAndCurrentSprint(sprintTaskID, sprintID).
		Select(`
            sprint_member_tasks.*,
            users.*,
            sprint_members.sprint_id,
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_members.member_id) AS total_points,
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_members.member_id, sprint_members.sprint_id) AS sprint_points,
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_members.member_id) AS total_time,
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_members.member_id, sprint_members.sprint_id) AS sprint_time`).
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
func (service SprintTaskService) GetMember(
	sprintMemberTask retroModels.SprintMemberTask,
	memberID uint, sprintID string) (member *retroSerializers.TaskMember, status int, err error) {
	db := service.DB
	member = new(retroSerializers.TaskMember)

	tempDB := service.smtForActiveAndCurrentSprint(fmt.Sprint(sprintMemberTask.SprintTaskID), sprintID).
		Where("sprint_members.member_id = ?", memberID).
		Select(`
            sprint_member_tasks.*,
            users.*, 
            sprint_members.sprint_id, 
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_tasks.task_id) AS total_points, 
            SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_tasks.task_id, sprint_members.sprint_id) AS sprint_points, 
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_tasks.task_id) AS total_time, 
            SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_tasks.task_id, sprint_members.sprint_id) AS sprint_time`).
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
func (service SprintTaskService) AddMember(
	sprintTaskID string,
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

	return service.GetMember(sprintMemberTask, sprintMember.MemberID, sprintID)
}

// UpdateTaskMember ...
func (service SprintTaskService) UpdateTaskMember(
	sprintTaskID string,
	retroID string,
	sprintID string,
	taskMemberData *retroSerializers.SprintTaskMemberUpdate) (*retroSerializers.TaskMember, int, error) {
	db := service.DB

	sprintMemberTask := retroModels.SprintMemberTask{}
	err := db.Model(&retroModels.SprintMemberTask{}).
		Where("task_id = ?", sprintTaskID).
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
func (service SprintTaskService) tasksForActiveAndCurrentSprint(retroID string, sprintID string, sprintTaskID *string) *gorm.DB {
	db := service.DB

	query := db.Model(retroModels.Task{}).
		Where("tasks.retrospective_id = ?", retroID).
		Scopes(retroModels.TaskJoinST, retroModels.STJoinSMT, retroModels.SMTJoinSM, retroModels.SMJoinSprint).
		Where("(sprints.status <> ? OR sprints.id = ?)", retroModels.DraftSprint, sprintID).
		Scopes(retroModels.NotDeletedSprint)
	if sprintTaskID != nil {
		sprintTaskFilter := db.Model(&retroModels.SprintTask{}).Where("id = ?", sprintTaskID).
			Select("task_id").QueryExpr()
		query = query.Where("sprint_tasks.task_id = (?)", sprintTaskFilter)
	}
	return query
}

// smtForActiveAndCurrentSprint ...
func (service SprintTaskService) smtForActiveAndCurrentSprint(sprintTaskID string, sprintID string) *gorm.DB {
	db := service.DB

	sprintTaskFilter := db.Model(&retroModels.SprintTask{}).Where("id = ?", sprintTaskID).
		Select("task_id").QueryExpr()

	return db.Model(retroModels.SprintMemberTask{}).
		Where("sprint_tasks.task_id = (?)", sprintTaskFilter).
		Scopes(
			retroModels.SMTJoinST,
			retroModels.STJoinTask,
			retroModels.SMTJoinSM,
			retroModels.SMJoinSprint,
			retroModels.SMJoinMember).
		Where("(sprints.status <> ? OR sprints.id = ?)", retroModels.DraftSprint, sprintID).
		Scopes(retroModels.NotDeletedSprint)
}
