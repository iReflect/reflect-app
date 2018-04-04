package services

// TODO Refactor this service and split into multiple services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/deckarep/golang-set"
	"github.com/gocraft/work"
	"github.com/jinzhu/gorm"

	"database/sql"
	"github.com/iReflect/reflect-app/apps/retrospective"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	taskTrackerSerializers "github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"github.com/iReflect/reflect-app/apps/timetracker"
	timeTrackerSerializers "github.com/iReflect/reflect-app/apps/timetracker/serializers"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
	"github.com/iReflect/reflect-app/workers"
	"net/http"
)

// SprintService ...
type SprintService struct {
	DB *gorm.DB
}

// DeleteSprint deletes the given sprint
func (service SprintService) DeleteSprint(sprintID string) (int, error) {
	db := service.DB

	// ToDo: Use batch deletes
	var sprint retroModels.Sprint
	err := db.Where("id = ?", sprintID).
		Where("status in (?)",
			[]retroModels.SprintStatus{retroModels.DraftSprint, retroModels.ActiveSprint}).
		Preload("SprintMembers.Tasks").
		Find(&sprint).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return http.StatusInternalServerError, errors.New("failed to get sprint")
	}
	tx := db.Begin()

	for _, sprintMember := range sprint.SprintMembers {
		for _, sprintMemberTask := range sprintMember.Tasks {
			if err := tx.Delete(&sprintMemberTask).Error; err != nil {
				tx.Rollback()
				utils.LogToSentry(err)
				return http.StatusInternalServerError, errors.New("sprint couldn't be deleted")
			}
		}
		if err := tx.Delete(&sprintMember).Error; err != nil {
			tx.Rollback()
			utils.LogToSentry(err)
			return http.StatusInternalServerError, errors.New("sprint couldn't be deleted")
		}
	}
	sprint.Status = retroModels.DeletedSprint
	err = tx.Exec("UPDATE sprints SET status = ? WHERE id = ?", retroModels.DeletedSprint, sprintID).Error
	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return http.StatusInternalServerError, errors.New("sprint couldn't be deleted")
	}
	err = tx.Commit().Error
	if err != nil {
		utils.LogToSentry(err)
		return http.StatusInternalServerError, errors.New("sprint couldn't be deleted")
	}
	return http.StatusNoContent, nil
}

// ActivateSprint activates the given sprint
func (service SprintService) ActivateSprint(sprintID string, retroID string) (int, error) {
	db := service.DB
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Where("status = ?", retroModels.DraftSprint).
		Find(&sprint).Error; err != nil {
		return http.StatusNotFound, nil
	}

	isValid, err := service.ValidateSprint(sprintID, retroID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if isValid {
		sprint.Status = retroModels.ActiveSprint
		if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
			return http.StatusInternalServerError, errors.New("sprint couldn't be activated")
		}
		return http.StatusNoContent, nil
	}
	return http.StatusBadRequest, errors.New("cannot activate an invalid draft sprint")
}

// FreezeSprint freezes the given sprint
func (service SprintService) FreezeSprint(sprintID string, retroID string) (int, error) {
	db := service.DB
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Where("status = ?", retroModels.ActiveSprint).
		Find(&sprint).Error; err != nil {
		return http.StatusNotFound, err
	}
	isValid, err := service.ValidateSprint(sprintID, retroID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if isValid {
		sprint.Status = retroModels.CompletedSprint
		if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
			return http.StatusInternalServerError, errors.New("sprint couldn't be frozen")
		}
		return http.StatusNoContent, nil
	}
	return http.StatusBadRequest, errors.New("can not freeze a invalid active sprint")
}

// Get return details of the given sprint
func (service SprintService) Get(sprintID string) (*retroSerializers.Sprint, int, error) {
	db := service.DB
	var sprint retroSerializers.Sprint

	err := db.Model(&retroModels.Sprint{}).Where("id = ?", sprintID).
		Preload("CreatedBy").
		First(&sprint).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
	}

	summary, status, err := service.GetSprintSummary(sprintID, sprint.RetrospectiveID)
	if err != nil {
		return nil, status, err
	}

	sprint.Summary = *summary

	err = db.Model(&retroModels.SprintSyncStatus{}).
		Where("sprint_id = ?", sprintID).
		Order("created_at DESC").
		Select("status").
		Row().Scan(&sprint.SyncStatus)

	if err != nil {
		if err != sql.ErrNoRows {
			utils.LogToSentry(err)
			return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
		}
		sprint.SyncStatus = int8(retroModels.NotSynced)
	}

	return &sprint, http.StatusOK, nil
}

// GetSprintSummary ...
func (service SprintService) GetSprintSummary(
	sprintID string,
	retroID uint) (*retroSerializers.SprintSummary, int, error) {
	db := service.DB

	var sprint retroModels.Sprint
	var summary retroSerializers.SprintSummary

	err := db.Model(&retroModels.Sprint{}).
		Where("id = ?", sprintID).
		Preload("Retrospective").
		First(&sprint).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint summary")
	}

	err = db.Model(&retroModels.Sprint{}).
		Scopes(retroModels.SprintJoinSM).
		Where("sprints.id = ?", sprintID).
		Select(`
            COUNT(*) AS member_count,
            SUM(allocation_percent) AS total_allocation,
            SUM(expectation_percent) AS total_expectation,
            SUM((? - vacations) * expectation_percent / 100.0 * allocation_percent / 100.0 * ?) AS target_sp,
            SUM(vacations) AS total_vacations,
            0 AS holidays`,
			utils.GetWorkingDaysBetweenTwoDates(*sprint.StartDate, *sprint.EndDate, true),
			sprint.Retrospective.StoryPointPerWeek/5).
		Scan(&summary).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint summary")
	}

	taskTypesSummary, status, err := service.GetSprintTaskSummary(sprintID, retroID)

	summary.TaskSummary = taskTypesSummary
	if err != nil {
		return nil, status, errors.New("failed to get sprint")
	}

	return &summary, http.StatusOK, nil
}

// GetSprintTaskSummary ...
func (service SprintService) GetSprintTaskSummary(
	sprintID string,
	retroID uint) (summary map[string]retroSerializers.SprintTaskSummary, status int, err error) {
	db := service.DB
	var retro retroModels.Retrospective

	err = db.Model(&retroModels.Retrospective{}).Where("id = ?", retroID).First(&retro).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("retrospective not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint summary")
	}

	taskTypes, err := tasktracker.GetTaskTypeMappings(retro.TaskProviderConfig)

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint summary")
	}
	summary = make(map[string]retroSerializers.SprintTaskSummary)

	for _, taskType := range tasktracker.TaskTypes {
		taskSummary, status, err := service.getSprintTaskTypeSummary(sprintID, taskTypes[taskType])
		if err != nil {
			return nil, status, err
		}
		summary[taskType] = *taskSummary
	}

	return summary, http.StatusOK, nil
}

func (service SprintService) getSprintTaskTypeSummary(
	sprintID string,
	taskTypes string) (*retroSerializers.SprintTaskSummary, int, error) {
	db := service.DB

	var summary retroSerializers.SprintTaskSummary

	taskList := db.Model(&retroModels.Sprint{}).
		Scopes(retroModels.SprintJoinSM, retroModels.SMJoinSMT, retroModels.SMTJoinTask).
		Where("sprints.id = ?", sprintID).
		Where("LOWER(tasks.type) IN (?)", taskTypes)

	doneTaskQuery := taskList.Where("sprints.start_date <= tasks.done_at").
		Where("sprints.end_date >= tasks.done_at").
		Select("COUNT(DISTINCT(tasks.id)) AS count, COALESCE(SUM(sprint_member_tasks.points_earned),0) AS points_earned").
		QueryExpr()

	taskQuery := taskList.Select(
		"COUNT(DISTINCT(tasks.id)) AS total_count, COALESCE(SUM(sprint_member_tasks.points_earned),0) AS total_points_earned").
		QueryExpr()
	err := db.Raw("SELECT * FROM (?) AS total CROSS JOIN (?) AS done", taskQuery, doneTaskQuery).
		Scan(&summary).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get task info for " + taskTypes)
	}

	return &summary, http.StatusOK, nil
}

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

	workers.Enqueuer.EnqueueUnique("sync_sprint_member_data",
		work.Q{"sprintMemberID": fmt.Sprint(sprintMember.ID)})

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

// SyncSprintData ...
func (service SprintService) SyncSprintData(sprintID string) (err error) {
	db := service.DB
	var sprint retroModels.Sprint
	err = db.Model(&retroModels.Sprint{}).
		Scopes(retroModels.NotDeletedSprint).
		Where("id = ?", sprintID).
		Preload("SprintMembers.Member").
		Preload("Retrospective").
		Find(&sprint).Error

	if err != nil {
		utils.LogToSentry(err)
		return err
	}

	if sprint.StartDate == nil || sprint.EndDate == nil {
		service.SetNotSynced(sprint.ID)
		return errors.New("sprint has no start/end date")
	}

	service.SetSyncing(sprint.ID)

	taskProviderConfig, err := tasktracker.DecryptTaskProviders(sprint.Retrospective.TaskProviderConfig)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	// TODO Restructure code-flow and document it to make it readable
	taskTrackerTaskKeySet, err := service.fetchAndUpdateTaskTrackerTask(sprint, taskProviderConfig)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	var timeTrackerTaskKeys []string
	var timeLogs []timeTrackerSerializers.TimeLog
	sprintMemberTimeLogs := map[uint][]timeTrackerSerializers.TimeLog{}
	for _, sprintMember := range sprint.SprintMembers {
		var memberTaskKeys []string
		memberTaskKeys, timeLogs, err = service.GetSprintMemberTimeTrackerData(sprintMember, sprint)
		sprintMemberTimeLogs[sprintMember.ID] = timeLogs
		timeTrackerTaskKeys = append(timeTrackerTaskKeys, memberTaskKeys...)
		if err != nil {
			utils.LogToSentry(err)
			service.SetSyncFailed(sprint.ID)
			return err
		}
	}

	insertedTimeTrackerTaskKeySet, err := service.fetchAndUpdateTimeTrackerTask(
		sprint.RetrospectiveID,
		taskProviderConfig,
		taskTrackerTaskKeySet,
		timeTrackerTaskKeys)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	err = service.updateMissingTimeTrackerTask(sprint.ID,
		sprint.RetrospectiveID,
		taskProviderConfig,
		timeTrackerTaskKeys,
		taskTrackerTaskKeySet,
		insertedTimeTrackerTaskKeySet)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}
	for _, sprintMember := range sprint.SprintMembers {
		err = service.updateSprintMemberTimeLog(
			sprint.ID,
			sprint.RetrospectiveID,
			sprintMember.ID,
			sprintMemberTimeLogs[sprintMember.ID])
		if err != nil {
			utils.LogToSentry(err)
			service.SetSyncFailed(sprint.ID)
			return err
		}

	}
	// ToDo: Store tickets not in SMT
	// Maybe a Join table ST

	service.SetSynced(sprint.ID)

	return nil
}

// SetNotSynced ...
func (service SprintService) SetNotSynced(sprintID uint) {
	db := service.DB
	db.Create(&retroModels.SprintSyncStatus{SprintID: sprintID, Status: retroModels.NotSynced})
}

// SetSyncing ...
func (service SprintService) SetSyncing(sprintID uint) {
	db := service.DB
	db.Create(&retroModels.SprintSyncStatus{SprintID: sprintID, Status: retroModels.Syncing})
}

// SetSyncFailed ...
func (service SprintService) SetSyncFailed(sprintID uint) {
	db := service.DB
	db.Create(&retroModels.SprintSyncStatus{SprintID: sprintID, Status: retroModels.SyncFailed})
}

// SetSynced ...
func (service SprintService) SetSynced(sprintID uint) {
	db := service.DB
	var sprint retroModels.Sprint
	db.Model(&retroModels.Sprint{}).
		Scopes(retroModels.NotDeletedSprint).
		Where("id = ?", sprintID).
		Find(&sprint)

	var currentTime time.Time
	currentTime = time.Now()
	sprint.LastSyncedAt = &currentTime

	db.Save(&sprint)

	db.Create(&retroModels.SprintSyncStatus{SprintID: sprint.ID, Status: retroModels.Synced})
}

// SyncSprintMemberData ...
func (service SprintService) SyncSprintMemberData(sprintMemberID string) (err error) {
	db := service.DB
	var sprintMember retroModels.SprintMember
	err = db.Model(&retroModels.SprintMember{}).
		Where("id = ?", sprintMemberID).
		Preload("Sprint").
		Preload("Member").
		Preload("Sprint.Retrospective").
		Find(&sprintMember).Error

	if err != nil {
		utils.LogToSentry(err)
		return err
	}

	sprint := sprintMember.Sprint

	if sprint.StartDate == nil || sprint.EndDate == nil {
		service.SetSyncFailed(sprint.ID)
		return errors.New("sprint has no start/end date")
	}

	service.SetSyncing(sprint.ID)

	taskProviderConfig, err := tasktracker.DecryptTaskProviders(sprint.Retrospective.TaskProviderConfig)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	taskTrackerTaskKeySet, err := service.fetchAndUpdateTaskTrackerTask(sprint, taskProviderConfig)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	var timeTrackerTaskKeys []string
	var timeLogs []timeTrackerSerializers.TimeLog
	timeTrackerTaskKeys, timeLogs, err = service.GetSprintMemberTimeTrackerData(sprintMember, sprint)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	insertedTimeTrackerTaskKeySet, err := service.fetchAndUpdateTimeTrackerTask(
		sprint.RetrospectiveID,
		taskProviderConfig,
		taskTrackerTaskKeySet,
		timeTrackerTaskKeys)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	err = service.updateMissingTimeTrackerTask(sprint.ID,
		sprint.RetrospectiveID,
		taskProviderConfig,
		timeTrackerTaskKeys,
		taskTrackerTaskKeySet,
		insertedTimeTrackerTaskKeySet)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	err = service.updateSprintMemberTimeLog(
		sprint.ID,
		sprint.RetrospectiveID,
		sprintMember.ID,
		timeLogs)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}
	// ToDo: Store tickets not in SMT
	// Maybe a Join table ST

	service.SetSynced(sprint.ID)

	return nil
}

// GetSprintsList ...
func (service SprintService) GetSprintsList(
	retrospectiveID string,
	userID uint) (sprints *retroSerializers.SprintsSerializer, status int, err error) {
	db := service.DB
	sprints = new(retroSerializers.SprintsSerializer)

	err = db.Model(&retroModels.Sprint{}).
		Where("retrospective_id = ?", retrospectiveID).
		Where("(sprints.status <> ? OR created_by_id = ?)", retroModels.DraftSprint, userID).
		Scopes(retroModels.NotDeletedSprint).
		Preload("CreatedBy").
		Order("end_date DESC, status, title, id").
		Find(&sprints.Sprints).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprints")
	}
	return sprints, http.StatusOK, nil
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

// Create creates a new sprint for the retro
func (service SprintService) Create(retroID string,
	sprintData retroSerializers.CreateSprintSerializer) (*retroSerializers.Sprint, int, error) {
	db := service.DB
	var err error
	var previousSprint retroModels.Sprint
	var sprint retroModels.Sprint
	var retro retroModels.Retrospective
	var sprintMember retroModels.SprintMember
	var iteratorLen int
	var previousSprintMembers []retroModels.SprintMember
	var teamMemberIDs []uint
	iteratorType := "member"

	err = db.Model(&retro).
		Where("id = ?", retroID).
		Find(&retro).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("retrospective not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get retrospective")
	}

	sprint.Title = sprintData.Title
	sprint.SprintID = sprintData.SprintID
	sprint.RetrospectiveID = retro.ID
	sprint.StartDate = sprintData.StartDate
	sprint.EndDate = sprintData.EndDate
	sprint.CreatedByID = sprintData.CreatedByID
	sprint.Status = retroModels.DraftSprint

	if sprint.SprintID != "" && sprint.StartDate == nil && sprint.EndDate == nil {

		taskProviderConfig, err := tasktracker.DecryptTaskProviders(retro.TaskProviderConfig)
		if err != nil {
			utils.LogToSentry(err)
			return nil, http.StatusInternalServerError,
				errors.New("failed to get task provider config. please contact admin")
		}

		connection := tasktracker.GetConnection(taskProviderConfig)
		if connection == nil {
			return nil, http.StatusInternalServerError, errors.New("invalid connection config")
		}

		var providerSprint *taskTrackerSerializers.Sprint

		// ToDo: The form will take task provider specific sprint ids in the future
		providerSprint = connection.GetSprint(sprintData.SprintID)
		if providerSprint != nil {
			if providerSprint.FromDate == nil || providerSprint.ToDate == nil {
				return nil, http.StatusUnprocessableEntity,
					errors.New("sprint doesn't have any start and/or end date. provide start date and end date " +
						"or set them in task provider")
			}
			sprint.StartDate = providerSprint.FromDate
			sprint.EndDate = providerSprint.ToDate
		}

		if providerSprint == nil {
			return nil, http.StatusUnprocessableEntity, errors.New("sprint id not found in task tracker")
		}
	}

	err = db.Model(&retroModels.Sprint{}).
		Where("sprints.status in (?)",
			[]retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}).
		Where("sprints.retrospective_id = ?", retro.ID).
		Order("sprints.end_date DESC, sprints.created_at DESC").
		First(&previousSprint).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, http.StatusInternalServerError, errors.New("failed to create sprint")
	} else if err == nil {
		if err = db.Model(&retroModels.SprintMember{}).
			Scopes(retroModels.SMJoinSprint, retroModels.SprintJoinRetro, retroModels.RetroJoinUserTeams).
			Where("sprint_members.sprint_id = ?", previousSprint.ID).
			Where("leaved_at IS NULL OR leaved_at >= ?", sprint.EndDate).
			Select("DISTINCT(sprint_members.*)").
			Find(&previousSprintMembers).Error; err != nil {
			return nil, http.StatusInternalServerError, errors.New("failed to create sprint")
		}
		iteratorLen = len(previousSprintMembers)
	}

	if iteratorLen < 1 {
		if err = db.Model(&userModels.UserTeam{}).
			Where("team_id = ?", retro.TeamID).
			Where("leaved_at IS NULL OR leaved_at >= ?", sprint.EndDate).
			Pluck("DISTINCT user_id", &teamMemberIDs).
			Error; err != nil {
			return nil, http.StatusInternalServerError, errors.New("failed to create sprint")
		}
		iteratorType = "memberID"
		iteratorLen = len(teamMemberIDs)

	}

	tx := db.Begin() // transaction begin

	if err = tx.Create(&sprint).Error; err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to create sprint")
	}

	for i := 0; i < iteratorLen; i++ {
		var userID uint
		allocationPercent := 100.00
		expectationPercent := 100.00
		if iteratorType == "member" {
			allocationPercent = previousSprintMembers[i].AllocationPercent
			expectationPercent = previousSprintMembers[i].ExpectationPercent
			userID = previousSprintMembers[i].MemberID
		} else {
			userID = teamMemberIDs[i]
		}
		sprintMember = retroModels.SprintMember{
			SprintID:           uint(sprint.ID),
			MemberID:           userID,
			Vacations:          0,
			Rating:             retrospective.DecentRating,
			AllocationPercent:  allocationPercent,
			ExpectationPercent: expectationPercent,
		}
		if err = tx.Create(&sprintMember).Error; err != nil {
			tx.Rollback()
			utils.LogToSentry(err)
			return nil, http.StatusInternalServerError, errors.New("failed to create sprint")
		}
	}

	err = tx.Commit().Error
	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to create sprint")
	}

	service.SetNotSynced(sprint.ID)
	workers.Enqueuer.EnqueueUnique("sync_sprint_data",
		work.Q{"sprintID": strconv.Itoa(int(sprint.ID)), "assignPoints": true})

	return service.Get(fmt.Sprint(sprint.ID))
}

// ValidateSprint validate the given sprint
func (service SprintService) ValidateSprint(sprintID string, retroID string) (bool, error) {
	db := service.DB
	query := `
		WITH constants (retro_id, sprint_id) AS (
		  VALUES (CAST (? AS INTEGER), CAST (? AS INTEGER))
		)
		SELECT COUNT(DISTINCT (t.*))
		FROM constants, (SELECT
                       tasks.id,
                       sm.sprint_id,
                       tasks.estimate,
                       SUM(smt.points_earned)
                       OVER (
                         PARTITION BY tasks.id )               AS total_points_earned,
                       SUM(smt.points_earned)
                       OVER (
                         PARTITION BY tasks.id, sm.sprint_id ) AS points_earned
                     FROM constants, tasks
                       JOIN sprint_member_tasks AS smt ON smt.task_id = tasks.id
                       JOIN sprint_members AS sm ON smt.sprint_member_id = sm.id
                       JOIN sprints ON sm.sprint_id = sprints.id
                     WHERE tasks.deleted_at IS NULL AND smt.deleted_at IS NULL AND sm."deleted_at" IS NULL AND
                            ((tasks.retrospective_id = constants.retro_id) AND
                            ((sprints.status <> ? OR sprints.id = constants.sprint_id)) AND
                            NOT (sprints.status = ?))) AS t
		WHERE t.sprint_id = constants.sprint_id AND (t.total_points_earned > t.estimate + 0.05)
	`
	var count struct {
		Count int
	}
	err := db.Raw(query, retroID, sprintID, retroModels.DraftSprint, retroModels.DeletedSprint).Scan(&count).Error
	if count.Count == 0 {
		return true, nil
	}
	if err != nil {
		utils.LogToSentry(err)
		return false, errors.New("error in fetching invalid sprint task lists")
	}
	return false, nil
}

// UpdateSprint updates the given sprint
func (service SprintService) UpdateSprint(sprintID string,
	sprintData retroSerializers.UpdateSprintSerializer) (*retroSerializers.Sprint, int, error) {
	db := service.DB
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
	}

	if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
		return nil, http.StatusInternalServerError, errors.New("sprint couldn't be updated")
	}

	return service.Get(sprintID)
}

// AssignPoints ...
func (service SprintService) AssignPoints(sprintID string) (err error) {
	fmt.Println("Assigning Points")
	db := service.DB
	var sprint retroModels.Sprint
	err = db.Model(&retroModels.Sprint{}).
		Scopes(retroModels.NotDeletedSprint).
		Where("id = ?", sprintID).
		Preload("SprintMembers").
		Preload("Retrospective").
		Find(&sprint).Error

	if err != nil {
		utils.LogToSentry(err)
		return err
	}

	service.SetSyncing(sprint.ID)

	dbs := db.Model(retroModels.SprintMemberTask{}).
		Scopes(retroModels.SMTJoinSM, retroModels.SMTJoinTask, retroModels.SMJoinSprint).
		Where("(sprints.status <> ? OR sprints.id = ?)", retroModels.DraftSprint, sprintID).
		Scopes(retroModels.NotDeletedSprint).
		Where("tasks.retrospective_id = ?", sprint.RetrospectiveID).
		Select(`
            sprint_member_tasks.*, 
            row_number() OVER (PARTITION BY sprint_member_tasks.task_id, sprint_members.sprint_id
                ORDER BY sprint_member_tasks.time_spent_minutes desc) AS time_spent_rank,
            sprint_members.sprint_id,
            (tasks.estimate - (SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_member_tasks.task_id)))
                AS remaining_points
        `).
		QueryExpr()

	err = db.Exec(`
    UPDATE sprint_member_tasks 
	    SET points_assigned = COALESCE(s1.remaining_points,0), 
            points_earned = COALESCE(s1.remaining_points, 0) 
        FROM (?) AS s1 
        WHERE s1.sprint_id = ? and time_spent_rank = 1 and sprint_member_tasks.id = s1.id
    `, dbs, sprintID).Error

	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	service.SetSynced(sprint.ID)

	return nil
}

// GetSprintMemberTimeTrackerData ...
func (service SprintService) GetSprintMemberTimeTrackerData(
	sprintMember retroModels.SprintMember,
	sprint retroModels.Sprint) ([]string, []timeTrackerSerializers.TimeLog, error) {

	timeLogs, err := timetracker.GetProjectTimeLogs(
		sprintMember.Member.TimeProviderConfig,
		sprint.Retrospective.ProjectName,
		*sprint.StartDate,
		*sprint.EndDate)

	if err != nil {
		utils.LogToSentry(err)
		return nil, nil, err
	}

	var ticketKeys []string
	for _, timeLog := range timeLogs {
		ticketKeys = append(ticketKeys, timeLog.TaskKey)
		if err != nil {
			utils.LogToSentry(err)
			return nil, nil, err
		}
	}
	return ticketKeys, timeLogs, nil
}

func (service SprintService) insertTimeTrackerTask(ticketKey string, retroID uint) (err error) {
	tx := service.DB.Begin()
	var task retroModels.Task
	err = tx.Where(retroModels.Task{RetrospectiveID: retroID, TrackerUniqueID: ticketKey}).
		Attrs(retroModels.Task{
			Key:         ticketKey,
			Summary:     "",
			Description: "",
			Type:        "",
			Priority:    "",
			Assignee:    "",
			Status:      ""}).FirstOrCreate(&task).Error
	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}
	err = tx.Where(retroModels.TaskKeyMap{TaskID: task.ID, Key: ticketKey}).
		FirstOrCreate(&retroModels.TaskKeyMap{}).Error

	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}
	tx.Commit()
	return nil
}

// changeTaskEstimates ...
func (service SprintService) changeTaskEstimates(task retroModels.Task, estimate float64) (err error) {
	tx := service.DB.Begin()

	switch {
	case task.Estimate > estimate:
		tx.Model(&task).UpdateColumn("estimate", estimate)
	case task.Estimate < estimate:
		tx.Model(&task).UpdateColumn("estimate", estimate)
		return nil
	default:
		return nil
	}

	fmt.Println("Rebalancing Points")
	activeAndFrozenSprintSMT := tx.Model(retroModels.SprintMemberTask{}).
		Scopes(retroModels.SMTJoinSM, retroModels.SMTJoinTask, retroModels.SMJoinSprint).
		Where("sprints.status <> ?", retroModels.DraftSprint).
		Scopes(retroModels.NotDeletedSprint).
		Where("tasks.id = ?", task.ID)

	dbs := activeAndFrozenSprintSMT.
		Where("sprint_member_tasks.points_earned != 0").
		Select(`
            sprint_member_tasks.*,
            sprint_members.sprint_id, 
			(tasks.estimate / (SUM(sprint_member_tasks.points_earned) 
                OVER (PARTITION BY sprint_member_tasks.task_id))) AS estimate_ratio`).
		QueryExpr()

	err = tx.Exec(`
        UPDATE 
            sprint_member_tasks
        SET
            points_earned = round((sprint_member_tasks.points_earned * estimate_ratio)::numeric,2)
		FROM
            (?) AS s1
		WHERE
            sprint_member_tasks.id = s1.id AND estimate_ratio < 1`, dbs).Error

	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}

	dbs = activeAndFrozenSprintSMT.
		Select(`
            DISTINCT(tasks.id)," +
            (tasks.estimate - (SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_member_tasks.task_id)))
                AS remaining_points`).
		QueryExpr()

	err = tx.Exec(`
        UPDATE sprint_member_tasks 
            SET points_earned = round(s2.target_earned::numeric,2)
        FROM (
            SELECT 
                DISTINCT(smt.id),
                s1.remaining_points,
                (SUM(smt.points_earned) OVER (PARTITION BY sm.sprint_id)) AS current_total,
                (s1.remaining_points * (points_earned / (SUM(smt.points_earned) OVER (PARTITION BY sm.sprint_id)))) AS target_earned
            FROM sprint_member_tasks 
                AS smt 
            JOIN sprint_members 
                AS sm 
                ON sm.id=smt.sprint_member_id AND sm.deleted_at IS NULL
            JOIN sprints 
                ON sprints.id=sm.sprint_id AND sprints.deleted_at IS NULL
            JOIN (?) AS s1 
                ON smt.task_id=s1.id 
            WHERE sprints.status = ? AND points_earned != 0
        ) AS s2 
        WHERE sprint_member_tasks.id = s2.id AND s2.current_total > s2.remaining_points
    `, dbs, retroModels.DraftSprint).Error

	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}
	tx.Commit()
	return nil
}

func (service SprintService) addOrUpdateTaskTrackerTask(
	ticket taskTrackerSerializers.Task,
	retroID uint,
	alternateTaskKey string) (err error) {
	tx := service.DB.Begin()

	var task retroModels.Task
	err = tx.Model(&retroModels.Task{}).
		Where(retroModels.Task{RetrospectiveID: retroID, TrackerUniqueID: ticket.TrackerUniqueID}).
		Assign(retroModels.Task{
			RetrospectiveID: retroID,
			TrackerUniqueID: ticket.TrackerUniqueID,
			Key:             ticket.Key,
			Summary:         ticket.Summary,
			Description:     ticket.Description,
			Type:            ticket.Type,
			Priority:        ticket.Priority,
			Assignee:        ticket.Assignee,
			Status:          ticket.Status,
		}).
		FirstOrCreate(&task).Error

	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}
	err = tx.Where(retroModels.TaskKeyMap{TaskID: task.ID, Key: ticket.Key}).
		FirstOrCreate(&retroModels.TaskKeyMap{}).Error

	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}

	if alternateTaskKey != "" {
		err = tx.Where(retroModels.TaskKeyMap{TaskID: task.ID, Key: alternateTaskKey}).
			FirstOrCreate(&retroModels.TaskKeyMap{}).Error

		if err != nil {
			tx.Rollback()
			utils.LogToSentry(err)
			return err
		}

	}
	if ticket.Estimate == nil {
		err = service.changeTaskEstimates(task, 0)
	} else {
		err = service.changeTaskEstimates(task, *ticket.Estimate)
	}
	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}
	tx.Commit()
	return nil
}

// fetchAndUpdateTaskTrackerTask ...
func (service SprintService) fetchAndUpdateTaskTrackerTask(
	sprint retroModels.Sprint,
	taskProviderConfig []byte) (mapset.Set, error) {
	taskTrackerTaskKeySet := mapset.NewSet()

	if sprint.SprintID != "" {
		tickets, err := tasktracker.GetSprintTaskList(taskProviderConfig, sprint.SprintID)
		if err != nil {
			utils.LogToSentry(err)
			return nil, err
		}

		for _, ticket := range tickets {
			err = service.addOrUpdateTaskTrackerTask(ticket, sprint.RetrospectiveID, "")
			if err != nil {
				utils.LogToSentry(err)
				return nil, err
			}
			taskTrackerTaskKeySet.Add(ticket.Key)
		}
	}
	return taskTrackerTaskKeySet, nil
}

// fetchAndUpdateTimeTrackerTask ...
func (service SprintService) fetchAndUpdateTimeTrackerTask(
	retroID uint,
	taskProviderConfig []byte,
	taskTrackerTaskKeySet mapset.Set,
	timeTrackerTaskKeys []string) (mapset.Set, error) {
	timeTrackerTaskKeySet := mapset.NewSetFromSlice(utils.StringSliceToInterfaceSlice(timeTrackerTaskKeys))
	missingTaskSet := timeTrackerTaskKeySet.Difference(taskTrackerTaskKeySet)
	missingTaskKeys := missingTaskSet.ToSlice()

	tickets, err := tasktracker.GetTaskList(taskProviderConfig, utils.InterfaceSliceToStringSlice(missingTaskKeys))
	if err != nil {
		utils.LogToSentry(err)
		return nil, err
	}

	timeTrackerTaskKeySet.Clear()
	for _, ticket := range tickets {
		err = service.addOrUpdateTaskTrackerTask(ticket, retroID, "")
		if err != nil {
			utils.LogToSentry(err)
			return nil, err
		}
		timeTrackerTaskKeySet.Add(ticket.Key)
	}
	return timeTrackerTaskKeySet, nil
}

// updateMissingTimeTrackerTask ...
func (service SprintService) updateMissingTimeTrackerTask(
	sprintID uint,
	retroID uint,
	taskProviderConfig []byte,
	timeTrackerTaskKeys []string,
	taskTrackerTaskKeySet mapset.Set,
	insertedTimeTrackerTaskKeySet mapset.Set) error {
	timeTrackerTaskKeySet := mapset.NewSetFromSlice(utils.StringSliceToInterfaceSlice(timeTrackerTaskKeys))
	missingTaskSet := timeTrackerTaskKeySet.Difference(insertedTimeTrackerTaskKeySet)
	missingTaskSet = missingTaskSet.Difference(taskTrackerTaskKeySet)
	missingTaskKeys := missingTaskSet.ToSlice()
	for _, taskKey := range missingTaskKeys {
		task, err := tasktracker.GetTaskDetails(taskProviderConfig, taskKey.(string))
		if err != nil {
			utils.LogToSentry(err)
			return err
		}

		if task != nil {
			err = service.addOrUpdateTaskTrackerTask(*task, retroID, taskKey.(string))
		} else {
			err = service.insertTimeTrackerTask(taskKey.(string), retroID)
		}
		if err != nil {
			utils.LogToSentry(err)
			return err
		}
	}
	return nil
}

func (service SprintService) addOrUpdateSMT(timeLog timeTrackerSerializers.TimeLog,
	sprintMemberID uint,
	retroID uint) (err error) {
	db := service.DB
	var sprintMemberTask retroModels.SprintMemberTask
	var task retroModels.Task
	err = db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_id = ?", sprintMemberID).
		Scopes(retroModels.SMTJoinTask, retroModels.TaskJoinTaskKeyMaps).
		Where("task_key_maps.key = ?", timeLog.TaskKey).
		Where("tasks.retrospective_id = ?", retroID).
		FirstOrInit(&sprintMemberTask).Error
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}

	err = db.Model(&retroModels.Task{}).
		Scopes(retroModels.TaskJoinTaskKeyMaps).
		Where("task_key_maps.key = ?", timeLog.TaskKey).
		Where("tasks.retrospective_id = ?", retroID).
		First(&task).Error
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}

	sprintMemberTask.SprintMemberID = sprintMemberID
	sprintMemberTask.TaskID = task.ID
	sprintMemberTask.TimeSpentMinutes += timeLog.Minutes

	return db.Save(&sprintMemberTask).Error
}

func (service SprintService) updateSprintMemberTimeLog(
	sprintID uint,
	retroID uint,
	sprintMemberID uint,
	timeLogs []timeTrackerSerializers.TimeLog) error {

	db := service.DB
	// Reset existing time_spent
	err := db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_id = ?", sprintMemberID).
		UpdateColumn("time_spent_minutes", 0).Error

	if err != nil {
		utils.LogToSentry(err)
		return err
	}
	for _, timeLog := range timeLogs {
		err = service.addOrUpdateSMT(timeLog, sprintMemberID, retroID)
		if err != nil {
			utils.LogToSentry(err)
			return err
		}
	}
	return nil
}
