package services

import (
	"errors"
	"strconv"

	"github.com/jinzhu/gorm"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	taskTrackerSerializers "github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	timeTrackerSerializers "github.com/iReflect/reflect-app/apps/timetracker/serializers"
)

// SprintService ...
type SprintService struct {
	DB *gorm.DB
}

// DeleteSprint deletes the given sprint
func (service SprintService) DeleteSprint(sprintID string) error {
	db := service.DB
	var sprint retroModels.Sprint
	if err := db.Where("id = ?", sprintID).
		Where("status in (?)", []retroModels.SprintStatus{retroModels.DraftSprint,
			retroModels.ActiveSprint}).
		Find(&sprint).Error; err != nil {
		return err
	}
	if rowsAffected := db.Delete(&sprint).RowsAffected; rowsAffected == 0 {
		return errors.New("sprint can't be deleted")
	}
	return nil
}

// ActivateSprint activates the given sprint
func (service SprintService) ActivateSprint(sprintID string) error {
	db := service.DB
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Where("status = ?", retroModels.DraftSprint).
		Find(&sprint).Error; err != nil {
		return err
	}

	sprint.Status = retroModels.ActiveSprint
	if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
		return errors.New("sprint couldn't be activated")
	}
	return nil
}

// Get return details of the given sprint
func (service SprintService) Get(sprintID string, userID uint) (*retrospectiveSerializers.Sprint, error) {
	db := service.DB
	var sprint retrospectiveSerializers.Sprint
	if err := db.Model(&retroModels.Sprint{}).
		Scopes(retroModels.NotDeletedSprint).
		Where("id = ?", sprintID).
		Preload("CreatedBy").
		Find(&sprint).Error; err != nil {
		return nil, err
	}
	return &sprint, nil
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
		return err
	}

	taskProviderConfig, err := tasktracker.DecryptTaskProviders(sprint.Retrospective.TaskProviderConfig)
	if err != nil {
		return err
	}

	tickets, err := tasktracker.GetSprintTaskList(taskProviderConfig, sprint.SprintID)
	if err != nil {
		return err
	}

	for _, ticket := range tickets {
		err = service.addOrUpdateTask(ticket, sprint.RetrospectiveID)
		if err != nil {
			return err
		}
	}

	for _, sprintMember := range sprint.SprintMembers {
		err = service.SyncSprintMemberData(strconv.Itoa(int(sprintMember.ID)))
		if err != nil {
			return err
		}
	}

	// ToDo: Store tickets not in SMT
	// Maybe a Join table ST

	return nil
}

// SyncSprintMemberData ...
func (service SprintService) SyncSprintMemberData(sprintMemberID string) (err error) {
	db := service.DB
	var sprintMember retroModels.SprintMember
	err = db.Model(&retroModels.SprintMember{}).
		Where("id = ?", sprintMemberID).
		Preload("Sprint.Retrospective").
		Find(&sprintMember).Error

	if err != nil {
		return err
	}

	// ToDo: Get tickets from TimeTracker
	timeLogs := []timeTrackerSerializers.TimeLog{}
	ticketIDs := []string{"IR-15", "IR-19"}

	taskProviderConfig, err := tasktracker.DecryptTaskProviders(sprintMember.Sprint.Retrospective.TaskProviderConfig)
	if err != nil {
		return err
	}

	tickets, err := tasktracker.GetTaskList(taskProviderConfig, ticketIDs)
	if err != nil {
		return err
	}

	for _, ticket := range tickets {
		err = service.addOrUpdateTask(ticket, sprintMember.Sprint.Retrospective.ID)
		if err != nil {
			return err
		}
	}

	// Reset existing time_spent
	err = db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_id = ?", sprintMemberID).
		UpdateColumn("time_spent_minutes", 0).Error

	if err != nil {
		return err
	}

	for _, timeLog := range timeLogs {
		err = service.addOrUpdateSMT(timeLog, sprintMember.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (service SprintService) addOrUpdateTask(ticket taskTrackerSerializers.Task, retroID uint) (err error) {
	// ToDo: Handle moved issues! ie ticket id changes
	db := service.DB
	var task retroModels.Task
	err = db.Model(&retroModels.Task{}).
		Where("retrospective_id = ?", retroID).
		Where("task_id = ?", ticket.ID).
		FirstOrInit(&task).Error
	if err != nil {
		return err
	}
	task.Summary = ticket.Summary
	task.TaskID = ticket.ID
	task.RetrospectiveID = retroID
	task.Type = ticket.Type
	task.Priority = ticket.Priority
	task.Estimate = ticket.Estimate
	task.Assignee = ticket.Assignee
	task.Status = ticket.Status

	return db.Save(&task).Error
}

func (service SprintService) addOrUpdateSMT(timeLog timeTrackerSerializers.TimeLog, sprintMemberID uint) (err error) {
	db := service.DB
	var sprintMemberTask retroModels.SprintMemberTask
	var task retroModels.Task
	err = db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_id = ?", sprintMemberID).
		Joins("tasks ON tasks.id=sprint_member_tasks.task_id").
		Where("tasks.task_id = ?", timeLog.TaskID).
		FirstOrInit(&sprintMemberTask).Error
	if err != nil {
		return err
	}

	err = db.Model(&retroModels.Task{}).
		Where("task_id = ?", timeLog.TaskID).
		First(&task).Error
	if err != nil {
		return err
	}

	sprintMemberTask.SprintMemberID = sprintMemberID
	sprintMemberTask.TaskID = task.ID
	sprintMemberTask.TimeSpentMinutes = timeLog.Minutes

	return db.Save(&sprintMemberTask).Error
}

// GetSprintsList ...
func (service SprintService) GetSprintsList(retrospectiveID string, userID uint) (sprints *retrospectiveSerializers.SprintsSerializer, err error) {
	db := service.DB
	sprints = new(retrospectiveSerializers.SprintsSerializer)

	err = db.Model(&retroModels.Sprint{}).
		Where("retrospective_id = ?", retrospectiveID).
		Where("status in (?) OR (status = (?) AND created_by_id = (?))", []retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}, retroModels.DraftSprint, userID).
		Preload("CreatedBy").
		Order("end_date desc").
		Scan(&sprints.Sprints).Error

	if err != nil {
		return nil, err
	}
	return sprints, nil
}
