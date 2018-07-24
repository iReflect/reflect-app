package services

// TODO Refactor this service and migrate to SprintSync service

import (
	"errors"
	"fmt"
	"time"

	"github.com/deckarep/golang-set"

	"github.com/gocraft/work"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	taskTrackerSerializers "github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"github.com/iReflect/reflect-app/apps/timetracker"
	timeTrackerSerializers "github.com/iReflect/reflect-app/apps/timetracker/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
	"github.com/iReflect/reflect-app/workers"
	"github.com/jinzhu/gorm"
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
		sprint.ID,
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

// QueueSprint ...
func (service SprintService) QueueSprint(sprintID uint, assignPoints bool) {
	db := service.DB
	workers.Enqueuer.EnqueueUnique("sync_sprint_data", work.Q{"sprintID": fmt.Sprint(sprintID), "assignPoints": assignPoints})
	db.Create(&retroModels.SprintSyncStatus{SprintID: sprintID, Status: retroModels.Queued})
}

// QueueSprintMember ...
func (service SprintService) QueueSprintMember(sprintID uint, sprintMemberID string) {
	db := service.DB
	workers.Enqueuer.EnqueueUnique("sync_sprint_member_data", work.Q{"sprintMemberID": sprintMemberID})
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
func (service SprintService) SyncSprintMemberData(sprintMemberID string) (err error) {
	db := service.DB
	var sprintMember retroModels.SprintMember
	err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.deleted_at IS NULL").
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
		sprint.ID,
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

// AssignPoints ...
func (service SprintService) AssignPoints(sprintID string) (err error) {
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

	stWithNoOrDifferingPointsSMTQuery := db.Model(retroModels.SprintTask{}).
		Scopes(retroModels.STLeftJoinSMT).
		Where("sprint_tasks.sprint_id = ?", sprintID).
		Where("sprint_member_tasks.id IS NULL").
		Or("sprint_member_tasks.points_earned <> sprint_member_tasks.points_assigned").
		Select("DISTINCT sprint_tasks.id").QueryExpr()

	dbs := db.Model(retroModels.SprintMemberTask{}).
		Where("sprint_member_tasks.deleted_at IS NULL").
		Scopes(retroModels.SMTJoinSM, retroModels.SMTJoinST, retroModels.STJoinTask, retroModels.SMJoinSprint).
		Where("(sprints.status <> ? OR sprints.id = ?)", retroModels.DraftSprint, sprintID).
		Not("sprint_tasks.id in (?)", stWithNoOrDifferingPointsSMTQuery).
		Scopes(retroModels.NotDeletedSprint).
		Where("tasks.retrospective_id = ?", sprint.RetrospectiveID).
		Select(`
            sprint_member_tasks.*, 
            sprint_members.sprint_id,
            (SUM(sprint_member_tasks.time_spent_minutes)
				OVER (PARTITION BY sprint_tasks.task_id, sprint_members.sprint_id)::numeric) AS total_time_spent,
            (tasks.estimate - (SUM(sprint_member_tasks.points_earned) OVER
				(PARTITION BY sprint_tasks.task_id)) + (SUM(sprint_member_tasks.points_earned) OVER
				(PARTITION BY sprint_tasks.id))) AS remaining_points
        `).
		QueryExpr()

	err = db.Exec(`
    UPDATE sprint_member_tasks
	    SET points_assigned = COALESCE(sprint_member_tasks.time_spent_minutes
				/ NULLIF(s1.total_time_spent, 0) * s1.remaining_points, 0),
            points_earned = COALESCE(sprint_member_tasks.time_spent_minutes
				/ NULLIF(s1.total_time_spent, 0) * s1.remaining_points, 0),
            updated_at = NOW()
        FROM (?) AS s1 
        WHERE s1.sprint_id = ? and sprint_member_tasks.id = s1.id
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

func (service SprintService) insertTimeTrackerTask(sprintID uint, ticketKey string, retroID uint) (err error) {
	tx := service.DB.Begin()
	var task retroModels.Task
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
	sprintID uint,
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

	err = tx.Where(retroModels.SprintTask{SprintID: sprintID, TaskID: task.ID}).
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
		err = service.addOrUpdateTaskTrackerTask(sprint.ID, ticket, sprint.RetrospectiveID, "")
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
	sprintID uint,
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
		err = service.addOrUpdateTaskTrackerTask(sprintID, ticket, retroID, "")
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
			err = service.addOrUpdateTaskTrackerTask(sprintID, *task, retroID, taskKey.(string))
		} else {
			err = service.insertTimeTrackerTask(sprintID, taskKey.(string), retroID)
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
