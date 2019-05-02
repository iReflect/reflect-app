package services

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"

	"strings"

	"github.com/iReflect/reflect-app/apps/retrospective"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	taskTrackerSerializers "github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/constants"
	customErrors "github.com/iReflect/reflect-app/libs"
	"github.com/iReflect/reflect-app/libs/utils"
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
func (service SprintService) DeleteSprint(sprintID string) (int, string, error) {
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
			responseError := constants.APIErrorMessages[constants.SprintNotFoundError]
			return http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.UnableToGetSprintError]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	// Check if the sprint is deleteable.
	isDeletableSprint, err := service.isSprintDeletable(sprint)
	if err == nil {
		if !isDeletableSprint {
			responseError := constants.APIErrorMessages[constants.DeleteSprintError]
			return http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
		}
	} else {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.SomethingWentWrong]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	// Transaction to delete the sprint starts
	tx := db.Begin()

	for _, sprintMember := range sprint.SprintMembers {
		for _, sprintMemberTask := range sprintMember.Tasks {
			if err := tx.Delete(&sprintMemberTask).Error; err != nil {
				tx.Rollback()
				utils.LogToSentry(err)
				responseError := constants.APIErrorMessages[constants.DeleteSprintError]
				return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
			}
		}
		if err := tx.Delete(&sprintMember).Error; err != nil {
			tx.Rollback()
			utils.LogToSentry(err)
			responseError := constants.APIErrorMessages[constants.DeleteSprintError]
			return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
		}
	}

	for _, sprintTask := range sprint.SprintTasks {
		if err := tx.Delete(&sprintTask).Error; err != nil {
			tx.Rollback()
			utils.LogToSentry(err)
			responseError := constants.APIErrorMessages[constants.DeleteSprintError]
			return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
		}

	}

	sprint.Status = retroModels.DeletedSprint
	err = tx.Exec("UPDATE sprints SET status = ?, updated_at = NOW() WHERE id = ?",
		retroModels.DeletedSprint,
		sprintID).Error
	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.DeleteSprintError]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	err = tx.Commit().Error
	if err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.DeleteSprintError]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	return http.StatusNoContent, "", nil
}

// ActivateSprint activates the given sprint
func (service SprintService) ActivateSprint(sprintID string, retroID string) (int, string, error) {
	db := service.DB
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Where("sprints.deleted_at IS NULL").
		Where("status = ?", retroModels.DraftSprint).
		Find(&sprint).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.SprintNotFoundError]
			return http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.UnableToGetSprintError]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	isValid, errorCode, err := service.ValidateSprint(sprintID, retroID)
	if err != nil {
		return http.StatusInternalServerError, errorCode, err
	}
	if isValid {
		sprint.Status = retroModels.ActiveSprint
		if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
			responseError := constants.APIErrorMessages[constants.ActivateSprintError]
			return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
		}
		service.QueueSprint(sprint.ID, true)
		return http.StatusNoContent, "", nil
	}
	responseError := constants.APIErrorMessages[constants.InvalidDraftSprintAcivationError]
	return http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
}

// FreezeSprint freezes the given sprint
func (service SprintService) FreezeSprint(sprintID string, retroID string) (int, string, error) {
	db := service.DB
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Where("sprints.deleted_at IS NULL").
		Where("status = ?", retroModels.ActiveSprint).
		Find(&sprint).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.SprintNotFoundError]
			return http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.UnableToGetSprintError]
		return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	isValid, errorCode, err := service.ValidateSprint(sprintID, retroID)
	if err != nil {
		return http.StatusInternalServerError, errorCode, err
	}

	if isValid {
		sprint.Status = retroModels.CompletedSprint
		if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
			responseError := constants.APIErrorMessages[constants.FrozenSprintError]
			return http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
		}
		return http.StatusNoContent, "", nil
	}
	responseError := constants.APIErrorMessages[constants.FreezeInvalidActiveSprintError]
	return http.StatusBadRequest, responseError.Code, errors.New(responseError.Message)
}

// Get return details of the given sprint
func (service SprintService) Get(sprintID string, userID uint, includeSprintSummary bool) (*retroSerializers.Sprint, int, string, error) {
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
			responseError := constants.APIErrorMessages[constants.SprintNotFoundError]
			return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.UnableToGetSprintError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
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
		summary, status, errorCode, err := service.GetSprintSummary(sprintID, sprint.RetrospectiveID)
		if err != nil {
			return nil, status, errorCode, err
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
			responseError := constants.APIErrorMessages[constants.UnableToGetSprintError]
			return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
		}
		sprint.SyncStatus = int8(retroModels.NotSynced)
	}

	sprint.SetEditable(userID)
	return &sprint, http.StatusOK, "", nil
}

// GetSprintSummary ...
func (service SprintService) GetSprintSummary(
	sprintID string,
	retroID uint) (*retroSerializers.SprintSummary, int, string, error) {
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
			responseError := constants.APIErrorMessages[constants.SprintNotFoundError]
			return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.GetSprintSummaryError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
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
		responseError := constants.APIErrorMessages[constants.GetSprintSummaryError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	taskTypesSummary, status, errorCode, err := service.GetSprintTaskSummary(sprintID, retroID)

	summary.TaskSummary = taskTypesSummary
	if err != nil {
		return nil, status, errorCode, err
	}

	return &summary, http.StatusOK, "", nil
}

// GetSprintTaskSummary ...
func (service SprintService) GetSprintTaskSummary(
	sprintID string,
	retroID uint) (summary map[string]retroSerializers.SprintTaskSummary, status int, errorCode string, err error) {
	db := service.DB
	var retro retroModels.Retrospective

	err = db.Model(&retroModels.Retrospective{}).
		Where("retrospectives.deleted_at IS NULL").
		Where("id = ?", retroID).First(&retro).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.RetrospectiveNotFoundError]
			return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.GetSprintSummaryError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	taskTypes, err := tasktracker.GetTaskTypeMappings(retro.TaskProviderConfig)

	if err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.GetSprintSummaryError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	summary = make(map[string]retroSerializers.SprintTaskSummary)

	for _, taskType := range tasktracker.TaskTypes {
		taskSummary, status, errorCode, err := service.getSprintTaskTypeSummary(sprintID, taskTypes[taskType])
		if err != nil {
			return nil, status, errorCode, err
		}
		summary[taskType] = *taskSummary
	}

	return summary, http.StatusOK, "", nil
}

func (service SprintService) getSprintTaskTypeSummary(
	sprintID string,
	taskTypes []string) (*retroSerializers.SprintTaskSummary, int, string, error) {
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
		responseError := constants.APIErrorMessages[constants.GetTaskDetailsError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message + "for" + strings.Join(taskTypes, ", "))
	}

	return &summary, http.StatusOK, "", nil
}

// GetSprintsList ...
func (service SprintService) GetSprintsList(retrospectiveID string, userID uint, perPage int, after string) (*retroSerializers.SprintsSerializer, int, string, error) {
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
			responseError := constants.APIErrorMessages[constants.GetSprintListError]
			return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
		}
		filterQuery = filterQuery.Where("sprints.end_date < ?", afterDate)
	}
	err := filterQuery.
		Order("end_date DESC, status, title, id").
		Limit(perPage).
		Find(&sprints.Sprints).Error

	if err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.GetSprintListError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	return sprints, http.StatusOK, "", nil
}

// Create creates a new sprint for the retro
func (service SprintService) Create(
	retroID string,
	userID uint,
	sprintData retroSerializers.CreateSprintSerializer) (*retroSerializers.Sprint, int, string, error) {
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
			responseError := constants.APIErrorMessages[constants.RetrospectiveNotFoundError]
			return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.RetrospectiveDetailsError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
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
			responseError := constants.APIErrorMessages[constants.GetTaskProviderConfigError]
			return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
		}

		connection := tasktracker.GetConnection(taskProviderConfig)
		if connection == nil {
			responseError := constants.APIErrorMessages[constants.InvalidConnectionConfigError]
			return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
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
			responseError := constants.APIErrorMessages[constants.SprintNotFoundInTaskTrackerError]
			return nil, http.StatusUnprocessableEntity, responseError.Code, errors.New(responseError.Message)
		}
	}
	if sprint.StartDate == nil || sprint.EndDate == nil {
		responseError := constants.APIErrorMessages[constants.SprintStartOrEndDateMissingError]
		return nil, http.StatusUnprocessableEntity, responseError.Code, errors.New(responseError.Message)
	}

	err = db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Where("sprints.status in (?)",
			[]retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}).
		Where("sprints.retrospective_id = ?", retro.ID).
		Order("sprints.end_date DESC, sprints.created_at DESC").
		First(&previousSprint).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.CreateSprintError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
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
			utils.LogToSentry(err)
			responseError := constants.APIErrorMessages[constants.CreateSprintError]
			return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
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
			responseError := constants.APIErrorMessages[constants.CreateSprintError]
			return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
		}
		iteratorType = "memberID"
		iteratorLen = len(teamMemberIDs)

	}

	tx := db.Begin() // transaction begin

	if err = tx.Create(&sprint).Error; err != nil {
		tx.Rollback()
		if customErrors.IsModelError(err) {
			return nil, http.StatusBadRequest, "", err
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.CreateSprintError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
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
			responseError := constants.APIErrorMessages[constants.CreateSprintError]
			return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
		}
	}

	err = tx.Commit().Error
	if err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.CreateSprintError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	service.SetNotSynced(sprint.ID)
	service.QueueSprint(sprint.ID, false)

	return service.Get(fmt.Sprint(sprint.ID), userID, true)
}

// ValidateSprint validate the given sprint
func (service SprintService) ValidateSprint(sprintID string, retroID string) (bool, string, error) {
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
		return true, "", nil
	}
	if err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.InvalidSprintTaskListError]
		return false, responseError.Code, errors.New(responseError.Message)
	}
	return false, "", nil
}

// UpdateSprint updates the given sprint
func (service SprintService) UpdateSprint(
	sprintID string,
	userID uint,
	sprintData retroSerializers.UpdateSprintSerializer) (*retroSerializers.Sprint, int, string, error) {
	db := service.DB
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Where("sprints.deleted_at IS NULL").
		Find(&sprint).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.SprintMemberNotFoundError]
			return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.UnableToGetSprintError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
		responseError := constants.APIErrorMessages[constants.UpdateSprintError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	return service.Get(sprintID, userID, true)
}
