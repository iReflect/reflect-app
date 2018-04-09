package services

// TODO Refactor this service and migrate to SprintMember service

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/retrospective"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	timeTrackerSerializers "github.com/iReflect/reflect-app/apps/timetracker/serializers"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
	"net/http"
)

// AddSprintMember ...
func (service SprintService) AddSprintMember(
	sprintID string,
	memberID uint) (*retroSerializers.SprintMemberSummary, int, error) {
	db := service.DB
	var sprintMember retroModels.SprintMember
	var sprint retroModels.Sprint

	err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprintID).
		Where("member_id = ?", memberID).
		Find(&retroModels.SprintMember{}).
		Error

	if err == nil {
		return nil, http.StatusBadRequest, errors.New("member already a part of the sprint")
	}

	err = db.Model(&retroModels.Sprint{}).
		Scopes(retroModels.SprintJoinRetro, retroModels.RetroJoinUserTeams).
		Where("user_teams.user_id=?", memberID).
		Where("sprints.id=?", sprintID).
		Preload("Retrospective").
		Find(&sprint).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusBadRequest, errors.New("member is not a part of the retrospective team")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to add member")
	}

	intSprintID, err := strconv.Atoi(sprintID)
	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusBadRequest, errors.New("failed to add member")
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
		return nil, http.StatusInternalServerError, errors.New("failed to add member")
	}

	service.QueueSprintMember(uint(intSprintID), fmt.Sprint(sprintMember.ID))

	sprintMemberSummary := new(retroSerializers.SprintMemberSummary)

	if err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprint.ID).
		Where("sprint_members.id = ?", sprintMember.ID).
		Scopes(retroModels.SMJoinMember).
		Select("DISTINCT sprint_members.*, users.*").
		Scan(&sprintMemberSummary).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("member not found in sprint")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get member summary")
	}

	sprintMemberSummary.ActualStoryPoint = 0
	sprintMemberSummary.SetExpectedStoryPoint(sprint, sprint.Retrospective)

	return sprintMemberSummary, http.StatusOK, nil
}

// RemoveSprintMember ...
func (service SprintService) RemoveSprintMember(sprintID string, memberID string) (int, error) {
	db := service.DB
	var sprintMember retroModels.SprintMember

	err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprintID).
		Where("id = ?", memberID).
		Preload("Tasks").
		Find(&sprintMember).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusNotFound, errors.New("sprint member not found")
		}
		utils.LogToSentry(err)
		return http.StatusInternalServerError, errors.New("failed to remove sprint member")
	}

	tx := db.Begin()
	for _, smt := range sprintMember.Tasks {
		err = tx.Delete(&smt).Error
		if err != nil {
			tx.Rollback()
			utils.LogToSentry(err)
			return http.StatusInternalServerError, errors.New("failed to remove sprint member")
		}
	}

	err = tx.Delete(&sprintMember).Error
	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return http.StatusInternalServerError, errors.New("failed to remove sprint member")
	}

	err = tx.Commit().Error

	if err != nil {
		utils.LogToSentry(err)
		return http.StatusInternalServerError, errors.New("failed to remove sprint member")
	}

	return http.StatusOK, nil
}

// GetSprintMembersSummary returns the sprint member summary list
func (service SprintService) GetSprintMembersSummary(
	sprintID string) (*retroSerializers.SprintMemberSummaryListSerializer, int, error) {
	db := service.DB
	sprintMemberSummaryList := new(retroSerializers.SprintMemberSummaryListSerializer)

	var sprint retroModels.Sprint
	err := db.Where("id = ?", sprintID).
		Preload("Retrospective").
		Find(&sprint).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
	}
	if err = db.Model(&retroModels.SprintMember{}).
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
		return nil, http.StatusInternalServerError, errors.New("failed to get member summary")
	}
	for _, sprintMemberSummary := range sprintMemberSummaryList.Members {
		sprintMemberSummary.SetExpectedStoryPoint(sprint, sprint.Retrospective)
	}
	return sprintMemberSummaryList, http.StatusOK, nil
}

// GetSprintMemberList returns the sprint member list
func (service SprintService) GetSprintMemberList(sprintID string) (sprintMemberList *userSerializers.MembersSerializer,
	status int, err error) {
	db := service.DB
	sprintMemberList = new(userSerializers.MembersSerializer)

	if err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprintID).
		Scopes(retroModels.SMJoinMember).
		Select("sprint_members.id, users.email, users.first_name, users.last_name, users.active").
		Order("users.first_name, users.last_name, users.id").
		Scan(&sprintMemberList.Members).
		Error; err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint members")
	}
	return sprintMemberList, http.StatusOK, nil
}

// UpdateSprintMember update the sprint member summary
func (service SprintService) UpdateSprintMember(sprintID string, sprintMemberID string,
	memberData retroSerializers.SprintMemberSummary) (*retroSerializers.SprintMemberSummary, int, error) {
	db := service.DB

	var sprintMember retroModels.SprintMember
	if err := db.Model(&retroModels.SprintMember{}).
		Where("id = ?", sprintMemberID).
		Where("sprint_id = ?", sprintID).
		Preload("Sprint.Retrospective").
		Find(&sprintMember).
		Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint member not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint member")
	}

	sprintMember.AllocationPercent = memberData.AllocationPercent
	sprintMember.ExpectationPercent = memberData.ExpectationPercent
	sprintMember.Vacations = memberData.Vacations
	sprintMember.Rating = retrospective.Rating(memberData.Rating)
	sprintMember.Comment = memberData.Comment

	if err := db.Save(&sprintMember).Error; err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to update sprint member")
	}

	if err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.id = ?", sprintMemberID).
		Scopes(retroModels.SMLeftJoinSMT).
		Select(`
            COALESCE(SUM(sprint_member_tasks.points_earned), 0) AS actual_story_point,
            COALESCE(SUM(sprint_member_tasks.time_spent_minutes), 0) AS total_time_spent_in_min`).
		Group("sprint_members.id").
		Row().
		Scan(&memberData.ActualStoryPoint, &memberData.TotalTimeSpentInMin); err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to update sprint member")
	}

	memberData.SetExpectedStoryPoint(sprintMember.Sprint, sprintMember.Sprint.Retrospective)

	return &memberData, http.StatusOK, nil
}

func (service SprintService) addOrUpdateSMT(timeLog timeTrackerSerializers.TimeLog,
	sprintMemberID uint,
	sprintID uint,
	retroID uint) (err error) {
	db := service.DB
	var sprintMemberTask retroModels.SprintMemberTask
	var sprintTask retroModels.SprintTask
	err = db.Model(&retroModels.SprintMemberTask{}).
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

	return db.Save(&sprintMemberTask).Error
}
