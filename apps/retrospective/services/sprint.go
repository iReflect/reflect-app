package services

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/retrospective"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	taskTrackerSerializers "github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	customErrors "github.com/iReflect/reflect-app/libs"
	"github.com/iReflect/reflect-app/libs/utils"
	"strings"
)

// SprintService ...
type SprintService struct {
	DB *gorm.DB
}

// Check if a given sprint is deletable or not
func (service SprintService) isSprintDeletable(sprint retroModels.Sprint) (bool, error) {
	db := service.DB
	// Check if the sprint is an active sprint sandwiched b/w a draft and a frozen sprint.
	if sprint.Status == retroModels.ActiveSprint {
		type SprintCount struct {
			DraftCount, FreezeCount int
		}
		var sprintCount SprintCount
		err := db.Model(&retroModels.Sprint{}).
			Where("sprints.deleted_at IS NULL").
			Scopes(retroModels.NotDeletedSprint).
			Where("retrospective_id = ?", sprint.RetrospectiveID).
			Select(
				"count(case when start_date > ? then id end) as draft_count, count(case when start_date < ? then id end) as freeze_count",
				sprint.EndDate, sprint.StartDate).Scan(&sprintCount).Error
		return !(sprintCount.DraftCount > 0 && sprintCount.FreezeCount > 0), err
	}
	return true, nil
}

// DeleteSprint deletes the given sprint
func (service SprintService) DeleteSprint(sprintID string) (int, error) {
	db := service.DB

	// ToDo: Use batch deletes
	var sprint retroModels.Sprint
	err := db.Where("id = ?", sprintID).
		Where("sprints.deleted_at IS NULL").
		Where("status in (?)",
			[]retroModels.SprintStatus{retroModels.DraftSprint, retroModels.ActiveSprint}).
		Preload("SprintMembers.Tasks").
		Preload("SprintTasks").
		Find(&sprint).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return http.StatusInternalServerError, errors.New("failed to get sprint")
	}

	// Check if the sprint is deleteable.
	isDeletableSprint, err := service.isSprintDeletable(sprint)
	if err == nil {
		if !isDeletableSprint {
			return http.StatusBadRequest, errors.New("this sprint can not be deleted.")
		}
	} else {
		utils.LogToSentry(err)
		return http.StatusInternalServerError, errors.New("something went wrong, please retry after some time.")
	}

	// Transaction to delete the sprint starts
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

	for _, sprintTask := range sprint.SprintTasks {
		if err := tx.Delete(&sprintTask).Error; err != nil {
			tx.Rollback()
			utils.LogToSentry(err)
			return http.StatusInternalServerError, errors.New("sprint couldn't be deleted")
		}

	}

	sprint.Status = retroModels.DeletedSprint
	err = tx.Exec("UPDATE sprints SET status = ?, updated_at = NOW() WHERE id = ?",
		retroModels.DeletedSprint,
		sprintID).Error
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
		Where("sprints.deleted_at IS NULL").
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
		service.QueueSprint(sprint.ID, true)
		return http.StatusNoContent, nil
	}
	return http.StatusBadRequest, errors.New("cannot activate an invalid draft sprint")
}

// FreezeSprint freezes the given sprint
func (service SprintService) FreezeSprint(sprintID string, retroID string) (int, error) {
	db := service.DB
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Where("sprints.deleted_at IS NULL").
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
func (service SprintService) Get(sprintID string, userID uint, includeSprintSummary bool) (*retroSerializers.Sprint, int, error) {
	db := service.DB
	var sprint retroSerializers.Sprint
	var currentSprint retroModels.Sprint

	err := db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Where("id = ?", sprintID).
		Preload("CreatedBy").
		Find(&currentSprint).
		First(&sprint).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, errors.New("sprint not found")
		}
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
	}

	// Check if the sprint is deleteable.
	sprint.Deletable = true
	isDeletableSprint, err := service.isSprintDeletable(currentSprint)
	if err == nil {
		sprint.Deletable = isDeletableSprint
	} else {
		utils.LogToSentry(err)
	}

	if includeSprintSummary {
		summary, status, err := service.GetSprintSummary(sprintID, sprint.RetrospectiveID)
		if err != nil {
			return nil, status, err
		}
		sprint.Summary = *summary
	}

	err = db.Model(&retroModels.SprintSyncStatus{}).
		Where("sprint_sync_statuses.deleted_at IS NULL").
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

	sprint.SetEditable(userID)
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
		Where("sprints.deleted_at IS NULL").
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
		Where("sprints.deleted_at IS NULL").
		Scopes(retroModels.SprintJoinSM).
		Where("sprints.id = ?", sprintID).
		Select(`
            COUNT(*) AS member_count,
            SUM(allocation_percent) AS total_allocation,
            SUM(expectation_percent) AS total_expectation,
            SUM((? - vacations) * expectation_percent / 100.0 * allocation_percent / 100.0 * ?) AS target_sp,
            SUM(vacations) AS total_vacations,
            0 AS holidays`,
			utils.GetWorkingDaysBetweenTwoDates(*sprint.StartDate, *sprint.EndDate),
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

	err = db.Model(&retroModels.Retrospective{}).
		Where("retrospectives.deleted_at IS NULL").
		Where("id = ?", retroID).First(&retro).Error
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
	taskTypes []string) (*retroSerializers.SprintTaskSummary, int, error) {
	db := service.DB

	var summary retroSerializers.SprintTaskSummary

	taskList := db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Scopes(retroModels.SprintJoinSM, retroModels.SMJoinSMT, retroModels.SMTJoinST, retroModels.STJoinTask).
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
		return nil, http.StatusInternalServerError, errors.New("failed to get task info for " + strings.Join(taskTypes, ", "))
	}

	return &summary, http.StatusOK, nil
}

// GetSprintsList ...
func (service SprintService) GetSprintsList(retrospectiveID string, userID uint, perPage int, after string) (*retroSerializers.SprintsSerializer, int, error) {
	db := service.DB
	sprints := new(retroSerializers.SprintsSerializer)

	filterQuery := db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Where("retrospective_id = ?", retrospectiveID).
		Scopes(retroModels.NotDeletedSprint).
		Preload("CreatedBy")

	if after != "" {
		afterDate, err := utils.ParseDateString(after)
		if err != nil {
			utils.LogToSentry(err)
			return nil, http.StatusInternalServerError, errors.New("failed to get sprints")
		}
		filterQuery = filterQuery.Where("sprints.end_date < ?", afterDate)
	}
	err := filterQuery.
		Order("end_date DESC, status, title, id").
		Limit(perPage).
		Find(&sprints.Sprints).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get sprints")
	}
	return sprints, http.StatusOK, nil
}

// Create creates a new sprint for the retro
func (service SprintService) Create(
	retroID string,
	userID uint,
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
		Where("retrospectives.deleted_at IS NULL").
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

	if sprint.SprintID != "" {

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
			if sprint.StartDate == nil {
				sprint.StartDate = providerSprint.FromDate
			}
			if sprint.EndDate == nil {
				sprint.EndDate = providerSprint.ToDate
			}
		} else {
			return nil, http.StatusUnprocessableEntity, errors.New("sprint not found in task tracker")
		}
	}
	if sprint.StartDate == nil || sprint.EndDate == nil {
		return nil, http.StatusUnprocessableEntity,
			errors.New("sprint doesn't have a start and/or end date, please provide the start date and end date " +
				"or set them in the task tracker")
	}

	err = db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Where("sprints.status in (?)",
			[]retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}).
		Where("sprints.retrospective_id = ?", retro.ID).
		Order("sprints.end_date DESC, sprints.created_at DESC").
		First(&previousSprint).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, http.StatusInternalServerError, errors.New("failed to create sprint")
	} else if err == nil {
		query := `SELECT DISTINCT (sprint_members.*)
            FROM "sprint_members"
            JOIN sprints ON sprint_members.sprint_id = sprints.id
            JOIN retrospectives ON sprints.retrospective_id = retrospectives.id
            JOIN user_teams ON retrospectives.team_id = user_teams.team_id
            WHERE "sprint_members"."deleted_at" IS NULL AND "sprint_members".member_id = user_teams.user_id AND (
            (sprints.deleted_at IS NULL) AND (retrospectives.deleted_at IS NULL) AND (user_teams.deleted_at IS NULL) AND
            (sprint_members.sprint_id = ?) AND
            ((leaved_at IS NULL OR leaved_at >= ?) AND joined_at <= ?))`
		if err = db.Raw(query, previousSprint.ID, sprint.StartDate, sprint.EndDate).Scan(&previousSprintMembers).Error; err != nil {
			return nil, http.StatusInternalServerError, errors.New("failed to create sprint")
		}
		iteratorLen = len(previousSprintMembers)
	}

	if iteratorLen < 1 {
		if err = db.Model(&userModels.UserTeam{}).
			Where("user_teams.deleted_at IS NULL").
			Where("team_id = ?", retro.TeamID).
			Where("(leaved_at IS NULL OR leaved_at >= ?) AND joined_at <= ? ", sprint.StartDate, sprint.EndDate).
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
		if customErrors.IsModelError(err) {
			return nil, http.StatusBadRequest, err
		}
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
	service.QueueSprint(sprint.ID, false)

	return service.Get(fmt.Sprint(sprint.ID), userID, true)
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
                     FROM constants, sprint_tasks AS st
                       JOIN sprint_member_tasks AS smt ON smt.sprint_task_id = st.id
                       JOIN tasks ON st.task_id = tasks.id
                       JOIN sprint_members AS sm ON smt.sprint_member_id = sm.id
                       JOIN sprints ON sm.sprint_id = sprints.id
                     WHERE tasks.deleted_at IS NULL AND smt.deleted_at IS NULL AND sm."deleted_at" IS NULL AND
							sprints."deleted_at" IS NULL AND
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
func (service SprintService) UpdateSprint(
	sprintID string,
	userID uint,
	sprintData retroSerializers.UpdateSprintSerializer) (*retroSerializers.Sprint, int, error) {
	db := service.DB
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Where("sprints.deleted_at IS NULL").
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

	return service.Get(sprintID, userID, true)
}
