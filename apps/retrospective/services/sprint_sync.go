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
	"github.com/iReflect/reflect-app/db"
	"github.com/iReflect/reflect-app/libs/utils"
	"github.com/iReflect/reflect-app/workers"
	"github.com/jinzhu/gorm"
)

// SyncSprintData ...
func (service SprintService) SyncSprintData(sprintID string) (err error) {
	db := db.DB
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

// QueueSprint ...
func (service SprintService) QueueSprint(sprintID uint, assignPoints bool) {
	db := db.DB
	workers.Enqueuer.EnqueueUnique("sync_sprint_data", work.Q{"sprintID": fmt.Sprint(sprintID), "assignPoints": assignPoints})
	db.Create(&retroModels.SprintSyncStatus{SprintID: sprintID, Status: retroModels.Queued})
}

// QueueSprintMember ...
func (service SprintService) QueueSprintMember(sprintID uint, sprintMemberID string) {
	workers.Enqueuer.EnqueueUnique("sync_sprint_member_data", work.Q{"sprintMemberID": sprintMemberID})
	db.DB.Create(&retroModels.SprintSyncStatus{SprintID: sprintID, Status: retroModels.Queued})
}

// SetNotSynced ...
func (service SprintService) SetNotSynced(sprintID uint) {
	db.DB.Create(&retroModels.SprintSyncStatus{SprintID: sprintID, Status: retroModels.NotSynced})
}

// SetSyncing ...
func (service SprintService) SetSyncing(sprintID uint) {
	db := db.DB
	db.Create(&retroModels.SprintSyncStatus{SprintID: sprintID, Status: retroModels.Syncing})
}

// SetSyncFailed ...
func (service SprintService) SetSyncFailed(sprintID uint) {
	db := db.DB
	db.Create(&retroModels.SprintSyncStatus{SprintID: sprintID, Status: retroModels.SyncFailed})
}

// SetSynced ...
func (service SprintService) SetSynced(sprintID uint) {
	db := db.DB
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
	db := db.DB
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

// AssignPoints ...
func (service SprintService) AssignPoints(sprintID string) (err error) {
	fmt.Println("Assigning Points")
	db := db.DB
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
	tx := db.DB.Begin()
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
func (service SprintService) changeTaskEstimates(tx *gorm.DB, task retroModels.Task, estimate float64) (err error) {
	txNotProvided := true
	if txNotProvided = tx == nil; txNotProvided {
		tx = db.DB.Begin()
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

	if txNotProvided {
		tx.Commit()
	}
	return nil
}

func (service SprintService) addOrUpdateTaskTrackerTask(
	ticket taskTrackerSerializers.Task,
	retroID uint,
	alternateTaskKey string) (err error) {
	tx := db.DB.Begin()

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

func (service SprintService) updateSprintMemberTimeLog(
	sprintID uint,
	retroID uint,
	sprintMemberID uint,
	timeLogs []timeTrackerSerializers.TimeLog) error {

	db := db.DB
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
