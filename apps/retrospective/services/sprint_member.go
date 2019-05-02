package services

// TODO Refactor this service and migrate to SprintMember service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"

	"net/http"

	"github.com/iReflect/reflect-app/apps/retrospective"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	timeTrackerSerializers "github.com/iReflect/reflect-app/apps/timetracker/serializers"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
	"github.com/iReflect/reflect-app/constants"
	"github.com/iReflect/reflect-app/libs/utils"
)

// AddSprintMember ...
func (service SprintService) AddSprintMember(
	sprintID string,
	memberID uint) (*retroSerializers.SprintMemberSummary, int, string, error) {
	db := service.DB
	var sprintMember retroModels.SprintMember
	var sprint retroModels.Sprint

	err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.deleted_at IS NULL").
		Where("sprint_id = ?", sprintID).
		Where("member_id = ?", memberID).
		Find(&retroModels.SprintMember{}).
		Error

	if err == nil {
		responseError := constants.APIErrorMessages[constants.MemberAlreadyInSprintError]
		return nil, http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
	}

	err = db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Scopes(retroModels.SprintJoinRetro, retroModels.RetroJoinUserTeams).
		Where("user_teams.user_id=?", memberID).
		Where("sprints.id=?", sprintID).
		Preload("Retrospective").
		Find(&sprint).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.NotRetroTeamMemberError]
			return nil, http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.UnableToAddMemberError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	intSprintID, err := strconv.Atoi(sprintID)
	if err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.UnableToAddMemberError]
		return nil, http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
	}

	sprintMember.SprintID = uint(intSprintID)
	sprintMember.MemberID = memberID
	sprintMember.Vacations = 0
	sprintMember.Rating = retrospective.DecentRating
	sprintMember.AllocationPercent = 100
	sprintMember.ExpectationPercent = 100

	err = db.Create(&sprintMember).Error
	if err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.UnableToAddMemberError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	service.QueueSprintMember(uint(intSprintID), fmt.Sprint(sprintMember.ID))

	sprintMemberSummary := new(retroSerializers.SprintMemberSummary)

	if err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.deleted_at IS NULL").
		Where("sprint_id = ?", sprint.ID).
		Where("sprint_members.id = ?", sprintMember.ID).
		Scopes(retroModels.SMJoinMember).
		Select("DISTINCT sprint_members.*, users.*").
		Scan(&sprintMemberSummary).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.MemberNotInSprintError]
			return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.GetMemberSummaryError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	sprintMemberSummary.ActualStoryPoint = 0
	sprintMemberSummary.SetExpectedStoryPoint(sprint, sprint.Retrospective)

	return sprintMemberSummary, http.StatusOK, "", nil
}

// RemoveSprintMember ...
func (service SprintService) RemoveSprintMember(sprintID string, memberID string) (int, string, error) {
	db := service.DB
	var sprintMember retroModels.SprintMember

	err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.deleted_at IS NULL").
		Where("sprint_id = ?", sprintID).
		Where("id = ?", memberID).
		Preload("Tasks").
		Find(&sprintMember).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.SprintMemberNotFoundError]
			return http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.RemoveSprintMemberError]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	tx := db.Begin()
	for _, smt := range sprintMember.Tasks {
		err = tx.Delete(&smt).Error
		if err != nil {
			tx.Rollback()
			utils.LogToSentry(err)
			responseError := constants.APIErrorMessages[constants.RemoveSprintMemberError]
			return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
		}
	}

	err = tx.Delete(&sprintMember).Error
	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.RemoveSprintMemberError]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	err = tx.Commit().Error

	if err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.RemoveSprintMemberError]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	return http.StatusOK, "", nil
}

// GetSprintMembersSummary returns the sprint member summary list
func (service SprintService) GetSprintMembersSummary(
	sprintID string) (*retroSerializers.SprintMemberSummaryListSerializer, int, string, error) {
	db := service.DB
	sprintMemberSummaryList := new(retroSerializers.SprintMemberSummaryListSerializer)

	var sprint retroModels.Sprint
	err := db.Where("id = ?", sprintID).
		Where("sprints.deleted_at IS NULL").
		Preload("Retrospective").
		Find(&sprint).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.SprintNotFoundError]
			return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.UnableToGetSprintError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	if err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.deleted_at IS NULL").
		Where("sprint_id = ?", sprint.ID).
		Scopes(retroModels.SMJoinMember, retroModels.SMLeftJoinSMT).
		Select(`
            DISTINCT sprint_members.*,
            users.*,
			SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_members.id) AS actual_story_point,
			SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_members.id) AS total_time_spent_in_min
		`).
		Order("users.first_name, users.last_name, users.id").
		Scan(&sprintMemberSummaryList.Members).
		Error; err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.GetMemberSummaryError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	for _, sprintMemberSummary := range sprintMemberSummaryList.Members {
		sprintMemberSummary.SetExpectedStoryPoint(sprint, sprint.Retrospective)
	}

	return sprintMemberSummaryList, http.StatusOK, "", nil
}

// GetSprintMemberList returns the sprint member list
func (service SprintService) GetSprintMemberList(sprintID string) (sprintMemberList *userSerializers.MembersSerializer,
	status int, errorCode string, err error) {
	db := service.DB
	sprintMemberList = new(userSerializers.MembersSerializer)

	if err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.deleted_at IS NULL").
		Where("sprint_id = ?", sprintID).
		Scopes(retroModels.SMJoinMember).
		Select("sprint_members.id, users.email, users.first_name, users.last_name, users.active").
		Order("users.first_name, users.last_name, users.id").
		Scan(&sprintMemberList.Members).
		Error; err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.GetSprintMemberListError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	return sprintMemberList, http.StatusOK, "", nil
}

// UpdateSprintMember update the sprint member summary
func (service SprintService) UpdateSprintMember(sprintID string, sprintMemberID string,
	memberData retroSerializers.SprintMemberUpdate) (*retroSerializers.SprintMemberSummary, int, string, error) {
	db := service.DB
	responseError := constants.APIErrorMessages[constants.UpdateSprintMemberError]
	return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	var sprintMember retroModels.SprintMember
	sprintMemberSummary := retroSerializers.SprintMemberSummary{}
	if err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.deleted_at IS NULL").
		Where("id = ?", sprintMemberID).
		Where("sprint_id = ?", sprintID).
		Preload("Sprint.Retrospective").
		Find(&sprintMember).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.SprintMemberNotFoundError]
			return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.GetSprintMemberError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	if memberData.AllocationPercent != nil {
		sprintMember.AllocationPercent = *memberData.AllocationPercent
	}
	if memberData.ExpectationPercent != nil {
		sprintMember.ExpectationPercent = *memberData.ExpectationPercent
	}
	if memberData.Vacations != nil {
		sprintMember.Vacations = *memberData.Vacations
	}
	if memberData.Rating != nil {
		sprintMember.Rating = retrospective.Rating(*memberData.Rating)
	}
	if memberData.Comment != nil {
		sprintMember.Comment = *memberData.Comment
	}

	if err := db.Save(&sprintMember).Error; err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.UpdateSprintMemberError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	if err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.deleted_at IS NULL").
		Where("sprint_members.id = ?", sprintMemberID).
		Scopes(retroModels.SMJoinMember, retroModels.SMLeftJoinSMT).
		Select(`
            DISTINCT sprint_members.*,
            users.*,
            COALESCE(SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_members.id), 0) AS actual_story_point,
            COALESCE(SUM(sprint_member_tasks.time_spent_minutes) OVER (PARTITION BY sprint_members.id), 0) AS total_time_spent_in_min`).
		Scan(&sprintMemberSummary).Error; err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.UpdateSprintMemberError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	sprintMemberSummary.SetExpectedStoryPoint(sprintMember.Sprint, sprintMember.Sprint.Retrospective)

	return &sprintMemberSummary, http.StatusOK, "", nil
}

func (service SprintService) addOrUpdateSMT(timeLog timeTrackerSerializers.TimeLog,
	sprintMemberID uint,
	sprintID uint,
	retroID uint) (err error) {
	db := service.DB
	var sprintMemberTask retroModels.SprintMemberTask
	var sprintTask retroModels.SprintTask
	err = db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_tasks.deleted_at IS NULL").
		Where("sprint_member_id = ?", sprintMemberID).
		Scopes(retroModels.SMTJoinST, retroModels.STJoinTask, retroModels.TaskJoinTaskKeyMaps).
		Where("task_key_maps.key = ?", timeLog.TaskKey).
		Where("tasks.retrospective_id = ?", retroID).
		FirstOrInit(&sprintMemberTask).Error
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}

	err = db.Model(&retroModels.SprintTask{}).
		Where("sprint_tasks.deleted_at IS NULL").
		Scopes(retroModels.STJoinTask, retroModels.TaskJoinTaskKeyMaps).
		Where("sprint_tasks.sprint_id = ?", sprintID).
		Where("task_key_maps.key = ?", timeLog.TaskKey).
		Where("tasks.retrospective_id = ?", retroID).
		First(&sprintTask).Error
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}

	sprintMemberTask.SprintMemberID = sprintMemberID
	sprintMemberTask.SprintTaskID = sprintTask.ID
	sprintMemberTask.TimeSpentMinutes = timeLog.Minutes

	return db.Set("smt:disable_validate", true).Save(&sprintMemberTask).Error
}
