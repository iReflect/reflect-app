package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gocraft/work"
	"github.com/jinzhu/gorm"

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
		Where("status in (?)", []retroModels.SprintStatus{retroModels.DraftSprint,
			retroModels.ActiveSprint}).
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
	if err := tx.Exec("UPDATE sprints set status=? where id=?", retroModels.DeletedSprint, sprintID).Error; err != nil {
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
	return http.StatusBadRequest, errors.New("can not activate a invalid draft sprint")
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

	err := db.Model(&retroModels.Sprint{}).Where("id = ?", sprintID).Preload("CreatedBy").First(&sprint).Error
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
		if err != gorm.ErrRecordNotFound {
			utils.LogToSentry(err)
			return nil, http.StatusInternalServerError, errors.New("failed to get sprint")
		}
		sprint.SyncStatus = int8(retroModels.NotSynced)
	}

	return &sprint, http.StatusOK, nil
}

// GetSprintSummary ...
func (service SprintService) GetSprintSummary(sprintID string, retroID uint) (*retroSerializers.SprintSummary, int, error) {
	db := service.DB

	var sprint retroModels.Sprint
	var summary retroSerializers.SprintSummary

	err := db.Model(&retroModels.Sprint{}).Where("id = ?", sprintID).Preload("Retrospective").First(&sprint).Error
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
		Select("COUNT(*) AS member_count, "+
			"SUM(allocation_percent) AS total_allocation, "+
			"SUM(expectation_percent) AS total_expectation, "+
			"SUM((? - vacations) * expectation_percent / 100.0 * allocation_percent / 100.0 * ?) AS target_sp, "+
			"SUM(vacations) AS total_vacations,"+
			"0 AS holidays", utils.GetWorkingDaysBetweenTwoDates(*sprint.StartDate, *sprint.EndDate, true), sprint.Retrospective.StoryPointPerWeek/5).
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
func (service SprintService) GetSprintTaskSummary(sprintID string, retroID uint) (summary map[string]retroSerializers.SprintTaskSummary, status int, err error) {
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

func (service SprintService) getSprintTaskTypeSummary(sprintID string, taskTypes string) (*retroSerializers.SprintTaskSummary, int, error) {
	db := service.DB

	var summary retroSerializers.SprintTaskSummary

	taskList := db.Model(&retroModels.Sprint{}).
		Scopes(retroModels.SprintJoinSM, retroModels.SMJoinSMT, retroModels.SMTJoinTask).
		Where("sprints.id = ?", sprintID).
		Where("LOWER(tasks.type) IN (?)", taskTypes)

	doneTaskQuery := taskList.Where("sprints.start_date <= tasks.done_at").
		Where("sprints.end_date >= tasks.done_at").
		Select("COUNT(*) AS count, COALESCE(SUM(sprint_member_tasks.points_earned),0) AS points_earned").
		QueryExpr()

	taskQuery := taskList.Select("COUNT(*) AS total_count, COALESCE(SUM(sprint_member_tasks.points_earned),0) AS total_points_earned").QueryExpr()
	err := db.Raw("SELECT * FROM (?) AS total CROSS JOIN (?) AS done", taskQuery, doneTaskQuery).Scan(&summary).Error

	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, errors.New("failed to get task info for " + taskTypes)
	}

	return &summary, http.StatusOK, nil
}

// AddSprintMember ...
func (service SprintService) AddSprintMember(sprintID string, memberID uint) (*retroSerializers.SprintMemberSummary, int, error) {
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
		Joins("JOIN retrospectives ON retrospectives.id=sprints.retrospective_id").
		Joins("JOIN user_teams ON retrospectives.team_id=user_teams.team_id").
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

	workers.Enqueuer.EnqueueUnique("sync_sprint_member_data", work.Q{"sprintMemberID": fmt.Sprint(sprintMember.ID)})

	sprintMemberSummary := new(retroSerializers.SprintMemberSummary)

	if err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprint.ID).
		Where("sprint_members.id = ?", sprintMember.ID).
		Joins("LEFT JOIN users ON users.id = sprint_members.member_id").
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
		Preload("SprintMembers").
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

	tickets, err := tasktracker.GetSprintTaskList(taskProviderConfig, sprint.SprintID)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	for _, ticket := range tickets {
		err = service.addOrUpdateTask(ticket, sprint.RetrospectiveID)
		if err != nil {
			service.SetSyncFailed(sprint.ID)
			return err
		}
	}

	for _, sprintMember := range sprint.SprintMembers {
		err = service.SyncSprintMemberData(strconv.Itoa(int(sprintMember.ID)), false)
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
func (service SprintService) SyncSprintMemberData(sprintMemberID string, independentRun bool) (err error) {
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
		if independentRun {
			service.SetSyncFailed(sprint.ID)
		}
		return errors.New("sprint has no start/end date")
	}

	if independentRun {
		service.SetSyncing(sprint.ID)
	}

	timeLogs, err := timetracker.GetProjectTimeLogs(sprintMember.Member.TimeProviderConfig, sprint.Retrospective.ProjectName, *sprint.StartDate, *sprint.EndDate)

	if err != nil {
		utils.LogToSentry(err)
		if independentRun {
			service.SetSyncFailed(sprint.ID)
		}
		return err
	}

	var ticketIDs []string
	for _, timeLog := range timeLogs {
		ticketIDs = append(ticketIDs, timeLog.TaskID)
		err = service.insertTask(timeLog.TaskID, sprintMember.Sprint.Retrospective.ID)
		if err != nil {
			utils.LogToSentry(err)
			return err
		}
	}

	taskProviderConfig, err := tasktracker.DecryptTaskProviders(sprintMember.Sprint.Retrospective.TaskProviderConfig)
	if err != nil {
		if independentRun {
			service.SetSyncFailed(sprint.ID)
		}
		return err
	}

	tickets, err := tasktracker.GetTaskList(taskProviderConfig, ticketIDs)
	if err != nil {
		utils.LogToSentry(err)
		if independentRun {
			service.SetSyncFailed(sprint.ID)
		}
		return err
	}

	for _, ticket := range tickets {
		err = service.addOrUpdateTask(ticket, sprintMember.Sprint.Retrospective.ID)
		if err != nil {
			utils.LogToSentry(err)
			if independentRun {
				service.SetSyncFailed(sprint.ID)
			}
			return err
		}
	}

	// Reset existing time_spent
	err = db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_id = ?", sprintMemberID).
		UpdateColumn("time_spent_minutes", 0).Error

	if err != nil {
		utils.LogToSentry(err)
		if independentRun {
			service.SetSyncFailed(sprint.ID)
		}
		return err
	}

	for _, timeLog := range timeLogs {
		err = service.addOrUpdateSMT(timeLog, sprintMember.ID, sprint.RetrospectiveID)
		if err != nil {
			utils.LogToSentry(err)
			if independentRun {
				service.SetSyncFailed(sprint.ID)
			}
			return err
		}
	}

	if independentRun {
		service.SetSynced(sprint.ID)
	}

	return nil
}

func (service SprintService) insertTask(ticketID string, retroID uint) (err error) {
	// ToDo: Handle moved issues! ie ticket id changes
	db := service.DB
	var task retroModels.Task
	err = db.Where(retroModels.Task{RetrospectiveID: retroID, TaskID: ticketID}).
		Attrs(retroModels.Task{Summary: "", Type: "", Priority: "", Assignee: "", Status: ""}).
		FirstOrCreate(&task).Error
	if err != nil {
		utils.LogToSentry(err)
		return err
	}
	return nil
}

func (service SprintService) addOrUpdateTask(ticket taskTrackerSerializers.Task, retroID uint) (err error) {
	db := service.DB
	var task retroModels.Task
	err = db.Where(retroModels.Task{RetrospectiveID: retroID, TaskID: ticket.ID}).
		Assign(retroModels.Task{
			Summary:  ticket.Summary,
			Type:     ticket.Type,
			Priority: ticket.Priority,
			Assignee: ticket.Assignee,
			Status:   ticket.Status,
		}).
		FirstOrCreate(&task).Error

	if err != nil {
		utils.LogToSentry(err)
		return err
	}

	if ticket.Estimate == nil {
		service.ChangeTaskEstimates(task, 0)
	} else {
		service.ChangeTaskEstimates(task, *ticket.Estimate)
	}

	return nil
}

func (service SprintService) addOrUpdateSMT(timeLog timeTrackerSerializers.TimeLog, sprintMemberID uint, retroID uint) (err error) {
	db := service.DB
	var sprintMemberTask retroModels.SprintMemberTask
	var task retroModels.Task
	err = db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_id = ?", sprintMemberID).
		Joins("JOIN tasks ON tasks.id=sprint_member_tasks.task_id").
		Where("tasks.task_id = ?", timeLog.TaskID).
		Where("tasks.retrospective_id = ?", retroID).
		FirstOrInit(&sprintMemberTask).Error
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}

	err = db.Model(&retroModels.Task{}).
		Where("task_id = ?", timeLog.TaskID).
		Where("tasks.retrospective_id = ?", retroID).
		First(&task).Error
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}

	sprintMemberTask.SprintMemberID = sprintMemberID
	sprintMemberTask.TaskID = task.ID
	sprintMemberTask.TimeSpentMinutes = timeLog.Minutes

	return db.Save(&sprintMemberTask).Error
}

// GetSprintsList ...
func (service SprintService) GetSprintsList(retrospectiveID string, userID uint) (sprints *retroSerializers.SprintsSerializer, status int, err error) {
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
func (service SprintService) GetSprintMembersSummary(sprintID string) (*retroSerializers.SprintMemberSummaryListSerializer, int, error) {
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
		Joins("JOIN users ON users.id = sprint_members.member_id").
		Joins("LEFT JOIN sprint_member_tasks AS smt ON smt.sprint_member_id = sprint_members.id").
		Select(`DISTINCT sprint_members.*,
                               users.*,
			                   SUM(smt.points_earned) over (PARTITION BY sprint_members.id) as actual_story_point,
			                   SUM(smt.time_spent_minutes) over (PARTITION BY sprint_members.id) as total_time_spent_in_min
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
func (service SprintService) GetSprintMemberList(sprintID string) (sprintMemberList *userSerializers.MembersSerializer, status int, err error) {
	db := service.DB
	sprintMemberList = new(userSerializers.MembersSerializer)

	if err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprintID).
		Joins("JOIN users ON users.id = sprint_members.member_id").
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
func (service SprintService) UpdateSprintMember(sprintID string, sprintMemberID string, memberData retroSerializers.SprintMemberSummary) (*retroSerializers.SprintMemberSummary, int, error) {
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
		Joins("LEFT JOIN sprint_member_tasks AS smt ON smt.sprint_member_id = sprint_members.id").
		Select("COALESCE(SUM(smt.points_earned), 0) as actual_story_point, COALESCE(SUM(smt.time_spent_minutes), 0) as total_time_spent_in_min").
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
func (service SprintService) Create(retroID string, sprintData retroSerializers.CreateSprintSerializer) (*retroSerializers.Sprint, int, error) {
	db := service.DB
	var err error
	var previousSprint retroModels.Sprint
	var sprint retroModels.Sprint
	var retro retroModels.Retrospective
	var sprintMember retroModels.SprintMember
	var iteratorLen int
	iteratorType := "member"
	var previousSprintMembers []retroModels.SprintMember
	var teamMemberIDs []uint

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
			return nil, http.StatusInternalServerError, errors.New("failed to get task provider config. please contact admin")
		}

		connections, err := tasktracker.GetConnections(taskProviderConfig)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		var providerSprint *taskTrackerSerializers.Sprint

		// ToDo: The form will take task provider specific sprint ids in the future
		for _, connection := range connections {
			providerSprint = connection.GetSprint(sprintData.SprintID)
			if providerSprint != nil {
				if providerSprint.FromDate == nil || providerSprint.ToDate == nil {
					return nil, http.StatusUnprocessableEntity, errors.New("sprint doesn't have any start and/or end date. provide start date and end date or set them in task provider")
				}
				sprint.StartDate = providerSprint.FromDate
				sprint.EndDate = providerSprint.ToDate
				break
			}
		}

		if providerSprint == nil {
			return nil, http.StatusUnprocessableEntity, errors.New("sprint id not found in task tracker")
		}
	}

	err = db.Model(&retroModels.Sprint{}).
		Where("sprints.status in (?)", []retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}).
		Where("sprints.retrospective_id = ?", retro.ID).
		Order("sprints.end_date DESC, sprints.created_at DESC").
		First(&previousSprint).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, http.StatusInternalServerError, errors.New("failed to create sprint")
	} else if err == nil {
		if err = db.Model(&retroModels.SprintMember{}).
			Joins("join sprints on sprint_members.sprint_id = sprints.id").
			Joins("join retrospectives on sprints.retrospective_id = retrospectives.id").
			Joins("join user_teams on user_teams.team_id = retrospectives.team_id").
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
	service.SetNotSynced(sprint.ID)

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
	workers.Enqueuer.EnqueueUnique("sync_sprint_data", work.Q{"sprintID": strconv.Itoa(int(sprint.ID)), "assignPoints": true})

	return service.Get(fmt.Sprint(sprint.ID))
}

// ValidateSprint validate the given sprint
func (service SprintService) ValidateSprint(sprintID string, retroID string) (bool, error) {
	db := service.DB
	inValidTasksCount := 0
	query := `
		WITH constants (retro_id, sprint_id) AS (
		  VALUES (?, ?)
		)
		SELECT DISTINCT (t.*)
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
	err := db.Raw(query, retroID, sprintID, retroModels.DraftSprint, retroModels.DeletedSprint).
		Count(&inValidTasksCount).Error
	if err != nil {
		utils.LogToSentry(err)
		return false, errors.New("error in fetching in-valid sprint task lists")
	}

	return inValidTasksCount == 0, nil

}

// UpdateSprint updates the given sprint
func (service SprintService) UpdateSprint(sprintID string, sprintData retroSerializers.UpdateSprintSerializer) (*retroSerializers.Sprint, int, error) {
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
		Joins("JOIN sprint_members AS sm ON sprint_member_tasks.sprint_member_id = sm.id").
		Joins("JOIN tasks ON tasks.id = sprint_member_tasks.task_id").
		Joins("JOIN sprints ON sm.sprint_id = sprints.id").
		Where("(sprints.status <> ? OR sprints.id = ?)", retroModels.DraftSprint, sprintID).
		Scopes(retroModels.NotDeletedSprint).
		Where("tasks.retrospective_id = ?", sprint.RetrospectiveID).
		Select(`sprint_member_tasks.*, 
                      row_number() over (PARTITION BY sprint_member_tasks.task_id, sm.sprint_id order by sprint_member_tasks.time_spent_minutes desc) as time_spent_rank,
                      sm.sprint_id,
                      (tasks.estimate - (SUM(sprint_member_tasks.points_earned) over (PARTITION BY sprint_member_tasks.task_id))) as remaining_points
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
	}

	service.SetSynced(sprint.ID)

	return nil
}

// ChangeTaskEstimates ...
func (service SprintService) ChangeTaskEstimates(task retroModels.Task, estimate float64) (err error) {
	db := service.DB

	switch {
	case task.Estimate > estimate:
		db.Model(&task).UpdateColumn("estimate", estimate)
	case task.Estimate < estimate:
		db.Model(&task).UpdateColumn("estimate", estimate)
		return nil
	default:
		return nil
	}

	fmt.Println("Rebalancing Points")
	activeAndFrozenSprintSMT := db.Model(retroModels.SprintMemberTask{}).
		Joins("JOIN sprint_members AS sm ON sprint_member_tasks.sprint_member_id = sm.id").
		Joins("JOIN tasks ON tasks.id = sprint_member_tasks.task_id").
		Joins("JOIN sprints ON sm.sprint_id = sprints.id").
		Where("sprints.status <> ?", retroModels.DraftSprint).
		Scopes(retroModels.NotDeletedSprint).
		Where("tasks.id = ?", task.ID)

	dbs := activeAndFrozenSprintSMT.
		Where("sprint_member_tasks.points_earned != 0").
		Select("sprint_member_tasks.*," +
			"sm.sprint_id, " +
			"(tasks.estimate / (SUM(sprint_member_tasks.points_earned) over (PARTITION BY sprint_member_tasks.task_id))) as estimate_ratio").
		QueryExpr()

	err = db.Exec("UPDATE sprint_member_tasks "+
		"SET points_earned = round((sprint_member_tasks.points_earned * estimate_ratio)::numeric,2) "+
		"FROM (?) AS s1 "+
		"WHERE sprint_member_tasks.id = s1.id AND estimate_ratio < 1", dbs).Error

	if err != nil {
		utils.LogToSentry(err)
		return err
	}

	dbs = activeAndFrozenSprintSMT.
		Select("DISTINCT(tasks.id)," +
			"(tasks.estimate - (SUM(sprint_member_tasks.points_earned) over (PARTITION BY sprint_member_tasks.task_id))) as remaining_points").
		QueryExpr()

	err = db.Exec(`
        UPDATE sprint_member_tasks 
            SET points_earned = round(s2.target_earned::numeric,2)
        FROM (
            SELECT 
                DISTINCT(smt.id),
                s1.remaining_points,
                (SUM(smt.points_earned) over (PARTITION BY sm.sprint_id)) as current_total,
                (s1.remaining_points * (points_earned / (SUM(smt.points_earned) over (PARTITION BY sm.sprint_id)))) as target_earned
            FROM sprint_member_tasks 
                AS smt 
            JOIN sprint_members 
                AS sm 
                ON sm.id=smt.sprint_member_id 
            JOIN sprints 
                ON sprints.id=sm.sprint_id 
            JOIN (?) AS s1 
                ON smt.task_id=s1.id 
            WHERE sprints.status = ? AND points_earned != 0
        ) AS s2 
        WHERE sprint_member_tasks.id = s2.id AND s2.current_total > s2.remaining_points
    `, dbs, retroModels.DraftSprint).Error

	if err != nil {
		utils.LogToSentry(err)
		return err
	}
	return nil
}
