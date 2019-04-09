package services

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/retrospective"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
)

// SprintTaskMemberService ...
type SprintTaskMemberService struct {
	DB *gorm.DB
}

// GetMembers ...
func (service SprintTaskMemberService) GetMembers(
	sprintTaskID string,
	retroID string,
	sprintID string) (members *retroSerializers.TaskMembersSerializer, status int, err error) {
	db := service.DB
	members = new(retroSerializers.TaskMembersSerializer)

	dbs := service.smtForCurrentAndPrevSprint(sprintTaskID, retroID, sprintID).
		Select(`
			DISTINCT ON (users.id)
			sprint_member_tasks.id, 
			sprint_member_tasks.created_at, 
			sprint_member_tasks.deleted_at, 
			sprint_member_tasks.updated_at, 
			sprint_member_tasks.role, 
			sprint_member_tasks.sprint_member_id, 
			sprint_member_tasks.sprint_task_id, 
			users.*, 
			sprints.end_date AS sprint_end_date, 
			sprint_members.sprint_id, 
			sprint_member_tasks.comment,
			sprint_member_tasks.rating,
			CASE WHEN (sprint_members.sprint_id = ?) THEN TRUE ELSE FALSE END AS current,
			SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_members.member_id) AS total_points,
			SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_members.member_id,sprint_members.sprint_id) AS sprint_points,
			SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_members.member_id) AS total_time,
			SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_members.member_id, sprint_members.sprint_id) AS sprint_time
		`, sprintID).
		Order("users.id DESC, sprints.end_date DESC").
		QueryExpr()

	err = db.Raw("SELECT smt.* FROM (?) AS smt", dbs).
		Order("smt.current DESC, smt.role, smt.first_name, smt.last_name").
		Scan(&members.Members).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get members")
	}

	return members, http.StatusOK, nil
}

// AddSprintActiveMember ...
func (service SprintTaskMemberService) AddSprintActiveMember(taskList *retroSerializers.SprintTasksSerializer,
	memberID uint,
	retroID string,
	sprintID string) *retroSerializers.SprintTasksSerializer {
	db := service.DB
	sprintMemberTask := retroModels.SprintMemberTask{}
	sprintMember := retroModels.SprintMember{}

	db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.deleted_at IS NULL").
		Where("sprint_id = ?", sprintID).
		Where("member_id = ?", memberID).
		Find(&sprintMember)

	sprintMemberTask.SprintMemberID = sprintMember.ID
	for _, task := range taskList.Tasks {
		sprintMemberTask.SprintTaskID = task.ID
		task.SprintActiveMemberTime = 0
		member, _, err := service.GetMember(sprintMemberTask, memberID, retroID, sprintID)

		if err == nil {
			task.SprintActiveMemberTime = member.SprintTime
		}
	}
	return taskList
}

// GetMember returns the task member summary of a task for a particular sprint member
func (service SprintTaskMemberService) GetMember(
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

	member.Current = true // Since the member task will always be a part of the sprint, current will always be True.

	return member, http.StatusOK, nil
}

// AddMember ...
func (service SprintTaskMemberService) AddMember(
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
func (service SprintTaskMemberService) UpdateTaskMember(
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

// smtForCurrentAndPrevSprint ...
func (service SprintTaskMemberService) smtForCurrentAndPrevSprint(sprintTaskID string, retroID string, sprintID string) *gorm.DB {
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
