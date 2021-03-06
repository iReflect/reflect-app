package services

// TODO Refactor this service and migrate to SprintSync service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/deckarep/golang-set"
	"github.com/gocraft/work"
	"github.com/jinzhu/gorm"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	taskTrackerSerializers "github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"github.com/iReflect/reflect-app/apps/timetracker"
	timeTrackerProviders "github.com/iReflect/reflect-app/apps/timetracker/providers"
	timeTrackerSerializers "github.com/iReflect/reflect-app/apps/timetracker/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
	"github.com/iReflect/reflect-app/workers"
)

// SyncSprintData ...
func (service SprintService) SyncSprintData(sprintID string) (err error) {
	db := service.DB
	var sprint retroModels.Sprint
	err = db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
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
	sprintMemberTimeLogs := map[uint][]timeTrackerSerializers.TimeLog{}

	timeTrackerTaskKeys, sprintMemberTimeLogs, err = service.GetTimeTrackerData(sprint, taskProviderConfig)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}
	insertedTimeTrackerTaskKeySet, err := service.fetchAndUpdateTimeTrackerTask(
		sprint,
		sprint.RetrospectiveID,
		taskProviderConfig,
		taskTrackerTaskKeySet,
		timeTrackerTaskKeys)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}
	err = service.updateMissingTimeTrackerTask(sprint,
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

// QueueSprint ...
func (service SprintService) QueueSprint(sprintID uint, assignPoints bool) {
	db := service.DB
	workers.Enqueuer.EnqueueUnique("sync_sprint_data", work.Q{"sprintID": fmt.Sprint(sprintID), "assignPoints": assignPoints})
	db.Create(&retroModels.SprintSyncStatus{SprintID: sprintID, Status: retroModels.Queued})
}

// QueueSprintMember ...
func (service SprintService) QueueSprintMember(sprintID uint, sprintMemberID string) {
	db := service.DB
	workers.Enqueuer.EnqueueUnique("sync_sprint_member_data",
		work.Q{"sprintID": sprintID, "sprintMemberID": sprintMemberID})
	db.Create(&retroModels.SprintSyncStatus{SprintID: sprintID, Status: retroModels.Queued})
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
		Where("sprints.deleted_at IS NULL").
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
func (service SprintService) SyncSprintMemberData(sprintID uint, sprintMemberID string) (err error) {
	db := service.DB

	service.SetSyncing(sprintID)

	var sprintMember retroModels.SprintMember
	err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.deleted_at IS NULL").
		Where("id = ?", sprintMemberID).
		Preload("Sprint").
		Preload("Member").
		Preload("Sprint.SprintMembers.Member").
		Preload("Sprint.Retrospective").
		Find(&sprintMember).Error

	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprintID)
		return err
	}

	sprint := sprintMember.Sprint

	if sprint.StartDate == nil || sprint.EndDate == nil {
		service.SetSyncFailed(sprint.ID)
		return errors.New("sprint has no start/end date")
	}

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
	timeProviderConfig := sprintMember.Member.TimeProviderConfig
	if sprint.Retrospective.TimeProviderName == timeTrackerProviders.TimeProviderJira {
		timeProviderConfig = taskProviderConfig
	}
	timeTrackerTaskKeys, timeLogs, err := service.GetSprintMemberSanitizedTimeTrackerData(taskProviderConfig, timeProviderConfig, sprint)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}
	var memberTimeLogs []timeTrackerSerializers.TimeLog
	for _, timeLog := range timeLogs {
		if sprintMember.Member.Email == timeLog.Email {
			memberTimeLogs = append(memberTimeLogs, timeLog)
		}
	}

	insertedTimeTrackerTaskKeySet, err := service.fetchAndUpdateTimeTrackerTask(
		sprint,
		sprint.RetrospectiveID,
		taskProviderConfig,
		taskTrackerTaskKeySet,
		timeTrackerTaskKeys)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	err = service.updateMissingTimeTrackerTask(sprint,
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
		memberTimeLogs)
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

// AssignPoints ...
func (service SprintService) AssignPoints(sprintID string, sprintTaskID *string) (err error) {
	fmt.Println("Assigning Points")
	db := service.DB
	var sprint retroModels.Sprint
	err = db.Model(&retroModels.Sprint{}).
		Where("id = ?", sprintID).
		Where("sprints.status = ?", retroModels.ActiveSprint).
		Where("sprints.deleted_at IS NULL").
		Preload("SprintMembers").
		Preload("Retrospective").
		Find(&sprint).Error

	if err != nil {
		utils.LogToSentry(err)
		return err
	}

	service.SetSyncing(sprint.ID)

	// sprintTaskToSkipPointsAllocation is the list of all the sprint tasks which can be skipped for the points allocation,
	// i.e., sprint tasks with no related SMTs or SMTs with differing points earned and points assigned value.
	sprintTaskToSkipPointsAllocation := db.Model(retroModels.SprintTask{}).
		Scopes(retroModels.STLeftJoinSMT).
		Where("sprint_tasks.sprint_id = ?", sprintID).
		Where("sprint_member_tasks.id IS NULL OR sprint_member_tasks.points_earned <> sprint_member_tasks.points_assigned").
		Select("DISTINCT sprint_tasks.id").QueryExpr()

	annotatedSMTExpr := db.Model(retroModels.SprintMemberTask{}).
		Where("sprint_member_tasks.deleted_at IS NULL").
		Scopes(retroModels.SMTJoinSM, retroModels.SMTJoinST, retroModels.STJoinTask, retroModels.SMJoinSprint).
		Where("(sprints.status <> ? OR sprints.id = ?)", retroModels.DraftSprint, sprintID).
		Not("sprint_tasks.id in (?)", sprintTaskToSkipPointsAllocation).
		Scopes(retroModels.NotDeletedSprint).
		Where("tasks.retrospective_id = ?", sprint.RetrospectiveID).
		Where("tasks.done_at IS NOT NULL").
		Select(`
            sprint_member_tasks.*, 
            sprint_members.sprint_id,
            (SUM(sprint_member_tasks.time_spent_minutes)
				OVER (PARTITION BY sprint_tasks.task_id, sprint_members.sprint_id)::numeric) AS sprint_task_total_time_spent,
            (tasks.estimate - (SUM(sprint_member_tasks.points_earned) OVER
				(PARTITION BY sprint_tasks.task_id)) + (SUM(sprint_member_tasks.points_earned) OVER
				(PARTITION BY sprint_tasks.id))) AS remaining_points
        `).QueryExpr()

	updateSQL := `UPDATE sprint_member_tasks
	    SET points_assigned = COALESCE(sprint_member_tasks.time_spent_minutes
				/ NULLIF(s1.sprint_task_total_time_spent, 0) * s1.remaining_points, 0),
            points_earned = COALESCE(sprint_member_tasks.time_spent_minutes
				/ NULLIF(s1.sprint_task_total_time_spent, 0) * s1.remaining_points, 0),
            updated_at = NOW()
        FROM (?) AS s1 
        WHERE s1.sprint_id = ? AND sprint_member_tasks.id = s1.id`

	sqlValues := []interface{}{annotatedSMTExpr, sprintID}

	if sprintTaskID != nil {
		updateSQL = fmt.Sprintf("%s AND s1.sprint_task_id = ?", updateSQL)
		sqlValues = append(sqlValues, *sprintTaskID)
	}

	err = db.Exec(updateSQL, sqlValues...).Error

	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return err
	}

	service.SetSynced(sprint.ID)

	return nil
}

// GetTimeTrackerData ...
func (service SprintService) GetTimeTrackerData(sprint retroModels.Sprint, taskProviderConfig []byte) (
	[]string,
	map[uint][]timeTrackerSerializers.TimeLog,
	error) {

	var timeTrackerTaskKeys []string
	var timeLogs []timeTrackerSerializers.TimeLog
	sprintMemberTimeLogs := map[uint][]timeTrackerSerializers.TimeLog{}
	var err error

	if sprint.Retrospective.TimeProviderName == timeTrackerProviders.TimeProviderJira {
		timeTrackerTaskKeys, timeLogs, err = service.GetSprintMemberTimeTrackerData(taskProviderConfig, sprint)
		if err != nil {
			utils.LogToSentry(err)
			service.SetSyncFailed(sprint.ID)
			return nil, nil, err
		}
		memberEmailTimeLogs := make(map[string][]timeTrackerSerializers.TimeLog)
		for _, timeLog := range timeLogs {
			memberEmailTimeLogs[timeLog.Email] = append(memberEmailTimeLogs[timeLog.Email], timeLog)
		}
		for _, sprintMember := range sprint.SprintMembers {
			sprintMemberTimeLogs[sprintMember.ID] = memberEmailTimeLogs[sprintMember.Member.Email]
		}
	} else {
		for _, sprintMember := range sprint.SprintMembers {
			var memberTaskKeys []string
			memberTaskKeys, timeLogs, err = service.GetSprintMemberSanitizedTimeTrackerData(taskProviderConfig, sprintMember.Member.TimeProviderConfig, sprint)
			if err != nil {
				utils.LogToSentry(err)
				service.SetSyncFailed(sprint.ID)
				return nil, nil, err
			}
			sprintMemberTimeLogs[sprintMember.ID] = timeLogs
			timeTrackerTaskKeys = append(timeTrackerTaskKeys, memberTaskKeys...)
		}
	}

	return timeTrackerTaskKeys, sprintMemberTimeLogs, nil
}

// GetSprintMemberSanitizedTimeTrackerData ...
func (service SprintService) GetSprintMemberSanitizedTimeTrackerData(
	taskProviderConfig []byte,
	timeTrackerConfig []byte,
	sprint retroModels.Sprint) ([]string, []timeTrackerSerializers.TimeLog, error) {

	taskKeys, timeLogs, err := service.GetSprintMemberTimeTrackerData(timeTrackerConfig, sprint)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return nil, nil, err
	}
	// Returns the map of uncleaned keys as map key and cleaned keys as map value
	sanitizedKeys, err := tasktracker.SanitizeTimeLogs(taskProviderConfig, taskKeys)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return nil, nil, err
	}

	timeLogsKeymap := make(map[string]int)
	sanitizedTimeLogs := make([]timeTrackerSerializers.TimeLog, len(taskKeys))
	var sanitizedMemberTaskKeys []string
	var sanitizedTimeLogsIndex = 0

	// To add the time of timelogs with the # in the key and same key without #
	// Also remove the repeated timelogs
	for _, timeLog := range timeLogs {
		var key = sanitizedKeys[timeLog.TaskKey]
		if value, ok := timeLogsKeymap[key]; ok {
			sanitizedTimeLogs[value].Minutes += timeLog.Minutes
		} else {
			timeLog.TaskKey = key
			sanitizedTimeLogs[sanitizedTimeLogsIndex] = timeLog
			timeLogsKeymap[key] = sanitizedTimeLogsIndex
			sanitizedMemberTaskKeys = append(sanitizedMemberTaskKeys, timeLog.TaskKey)
			sanitizedTimeLogsIndex++
		}
	}
	return sanitizedMemberTaskKeys, sanitizedTimeLogs, nil
}

// GetSprintMemberTimeTrackerData ...
func (service SprintService) GetSprintMemberTimeTrackerData(
	timeTrackerConfig []byte,
	sprint retroModels.Sprint) ([]string, []timeTrackerSerializers.TimeLog, error) {

	timeLogs, err := timetracker.GetProjectTimeLogs(
		timeTrackerConfig,
		sprint.Retrospective.ProjectName,
		*sprint.StartDate,
		*sprint.EndDate)
	if err != nil {
		utils.LogToSentry(err)
		service.SetSyncFailed(sprint.ID)
		return nil, nil, err
	}
	sprintMemberEmailMap := make(map[string]uint)
	for _, member := range sprint.SprintMembers {
		sprintMemberEmailMap[member.Member.Email] = member.MemberID
	}
	var ticketKeys []string
	for _, timeLog := range timeLogs {
		if _, exists := sprintMemberEmailMap[timeLog.Email]; exists {
			ticketKeys = append(ticketKeys, timeLog.TaskKey)
		}
	}
	return ticketKeys, timeLogs, nil
}

func (service SprintService) insertTimeTrackerTask(sprintID uint, ticketKey string, retroID uint) (err error) {
	tx := service.DB.Begin()
	var task retroModels.Task
	var sprint retroModels.Sprint

	err = service.DB.Model(retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Where("id = ?", sprintID).
		Find(&sprint).
		Error
	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}
	err = tx.Where(retroModels.Task{RetrospectiveID: retroID, TrackerUniqueID: ticketKey}).
		Where("tasks.deleted_at IS NULL").
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
	err = tx.Model(&retroModels.Task{}).
		Where("id = ?", task.ID).
		Updates(map[string]interface{}{"done_at": sprint.EndDate, "resolution": retroModels.DoneResolution}).Error

	if err != nil {
		utils.LogToSentry(err)
		return err
	}

	err = tx.Where(retroModels.TaskKeyMap{TaskID: task.ID, Key: ticketKey}).
		Where("task_key_maps.deleted_at IS NULL").
		FirstOrCreate(&retroModels.TaskKeyMap{}).Error

	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}

	err = tx.Where(retroModels.SprintTask{SprintID: sprintID, TaskID: task.ID}).
		Where("sprint_tasks.deleted_at IS NULL").
		FirstOrCreate(&retroModels.SprintTask{}).Error
	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}

	tx.Commit()
	return nil
}

// changeTaskEstimates ...
func (service SprintService) changeTaskEstimates(tx *gorm.DB, task retroModels.Task, estimate float64) (err error) {
	txNotProvided := true
	if txNotProvided = tx == nil; txNotProvided {
		tx = service.DB.Begin()
	}

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
		Where("sprint_member_tasks.deleted_at IS NULL").
		Scopes(retroModels.SMTJoinSM, retroModels.SMTJoinST, retroModels.STJoinTask, retroModels.SMJoinSprint).
		Where("sprints.status <> ?", retroModels.DraftSprint).
		Scopes(retroModels.NotDeletedSprint).
		Where("tasks.id = ?", task.ID)

	dbs := activeAndFrozenSprintSMT.
		Where("sprint_member_tasks.points_earned != 0").
		Select(`
            sprint_member_tasks.*,
            sprint_members.sprint_id, 
			(tasks.estimate / (SUM(sprint_member_tasks.points_earned) 
                OVER (PARTITION BY sprint_tasks.task_id))) AS estimate_ratio`).
		QueryExpr()

	err = tx.Exec(`
        UPDATE 
            sprint_member_tasks
        SET
            points_earned = round((sprint_member_tasks.points_earned * estimate_ratio)::numeric,2),
            updated_at = NOW()
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
            DISTINCT(tasks.id),
            (tasks.estimate - (SUM(sprint_member_tasks.points_earned) OVER (PARTITION BY sprint_tasks.task_id)))
                AS remaining_points`).
		QueryExpr()

	err = tx.Exec(`
        UPDATE sprint_member_tasks 
            SET points_earned = round(s2.target_earned::numeric,2),
            updated_at = NOW()
        FROM (
            SELECT 
                DISTINCT(smt.id),
                s1.remaining_points,
                (SUM(smt.points_earned) OVER (PARTITION BY sm.sprint_id)) AS current_total,
                (s1.remaining_points * (points_earned / (SUM(smt.points_earned) OVER (PARTITION BY sm.sprint_id)))) AS target_earned
            FROM sprint_member_tasks 
                AS smt 
            JOIN sprint_tasks
                AS st 
                ON st.id=smt.sprint_task_id AND st.deleted_at IS NULL
            JOIN sprint_members 
                AS sm 
                ON sm.id=smt.sprint_member_id AND sm.deleted_at IS NULL
            JOIN sprints 
                ON sprints.id=sm.sprint_id AND sprints.deleted_at IS NULL
            JOIN (?) AS s1 
                ON st.task_id=s1.id 
            WHERE sprints.status = ? AND points_earned != 0
        ) AS s2 
        WHERE sprint_member_tasks.id = s2.id AND s2.current_total > s2.remaining_points
    `, dbs, retroModels.DraftSprint).Error

	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}

	if txNotProvided {
		tx.Commit()
	}
	return nil
}

func (service SprintService) addOrUpdateTaskTrackerTask(
	sprint retroModels.Sprint,
	ticket taskTrackerSerializers.Task,
	retroID uint,
	alternateTaskKey string) (err error) {

	tx := service.DB.Begin()
	var task retroModels.Task

	err = tx.Model(&retroModels.Task{}).
		Where("tasks.deleted_at IS NULL").
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
			IsTrackerTask:   true,
		}).
		FirstOrCreate(&task).Error

	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}

	statusMap, err := tasktracker.GetStatusMapping(sprint.Retrospective.TaskProviderConfig)
	if err != nil {
		utils.LogToSentry(err)
		return errors.New("failed to fetch status mapping")
	}

	if len(statusMap[tasktracker.DoneStatus]) != 0 {
		for _, status := range statusMap[tasktracker.DoneStatus] {
			if strings.ToLower(ticket.Status) == status {
				err = tx.Model(&retroModels.Task{}).
					Where("id = ?", task.ID).
					Update("DoneAt", sprint.EndDate).Error

				if err != nil {
					utils.LogToSentry(err)
					return err
				}
				err = tx.Model(&retroModels.Task{}).
					Where("id = ?", task.ID).
					Where("resolution = ?", retroModels.TaskNotDoneResolution).
					Update("resolution", retroModels.DoneResolution).Error

				if err != nil {
					utils.LogToSentry(err)
					return err
				}
				break
			}
		}
	}
	err = tx.Where(retroModels.TaskKeyMap{TaskID: task.ID, Key: ticket.Key}).
		Where("task_key_maps.deleted_at IS NULL").
		FirstOrCreate(&retroModels.TaskKeyMap{}).Error

	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}

	if alternateTaskKey != "" {
		err = tx.Where(retroModels.TaskKeyMap{TaskID: task.ID, Key: alternateTaskKey}).
			Where("task_key_maps.deleted_at IS NULL").
			FirstOrCreate(&retroModels.TaskKeyMap{}).Error

		if err != nil {
			tx.Rollback()
			utils.LogToSentry(err)
			return err
		}

	}

	err = tx.Where(retroModels.SprintTask{SprintID: sprint.ID, TaskID: task.ID}).
		Where("sprint_tasks.deleted_at IS NULL").
		FirstOrCreate(&retroModels.SprintTask{}).Error

	if err != nil {
		tx.Rollback()
		utils.LogToSentry(err)
		return err
	}

	if ticket.Estimate == nil {
		err = service.changeTaskEstimates(tx, task, 0)
	} else {
		err = service.changeTaskEstimates(tx, task, *ticket.Estimate)
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

	tickets, err := tasktracker.GetSprintTaskList(
		taskProviderConfig,
		taskTrackerSerializers.Sprint{
			ID:       sprint.SprintID,
			FromDate: sprint.StartDate,
			ToDate:   sprint.EndDate,
		},
	)
	if err != nil {
		utils.LogToSentry(err)
		return nil, err
	}

	for _, ticket := range tickets {
		err = service.addOrUpdateTaskTrackerTask(sprint, ticket, sprint.RetrospectiveID, "")
		if err != nil {
			utils.LogToSentry(err)
			return nil, err
		}
		taskTrackerTaskKeySet.Add(ticket.Key)
	}
	return taskTrackerTaskKeySet, nil
}

// fetchAndUpdateTimeTrackerTask ...
func (service SprintService) fetchAndUpdateTimeTrackerTask(
	sprint retroModels.Sprint,
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
		err = service.addOrUpdateTaskTrackerTask(sprint, ticket, retroID, "")
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
	sprint retroModels.Sprint,
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
			err = service.addOrUpdateTaskTrackerTask(sprint, *task, retroID, taskKey.(string))
		} else {
			err = service.insertTimeTrackerTask(sprint.ID, taskKey.(string), retroID)
		}
		if err != nil {
			utils.LogToSentry(err)
			return err
		}
	}
	return nil
}

func (service SprintService) updateSprintMemberTimeLog(
	sprintID uint,
	retroID uint,
	sprintMemberID uint,
	timeLogs []timeTrackerSerializers.TimeLog) error {

	db := service.DB
	// Reset existing time_spent
	err := db.Exec("UPDATE sprint_member_tasks SET time_spent_minutes=0 WHERE sprint_member_id = ?", sprintMemberID).Error

	if err != nil {
		utils.LogToSentry(err)
		return err
	}
	for _, timeLog := range timeLogs {
		err = service.addOrUpdateSMT(timeLog, sprintMemberID, sprintID, retroID)
		if err != nil {
			utils.LogToSentry(err)
			return err
		}
	}
	return nil
}
