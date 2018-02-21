package services

import (
	"errors"
	"math"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/gocraft/work"
	"github.com/iReflect/reflect-app/apps/retrospective"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	taskTrackerSerializers "github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	timeTrackerSerializers "github.com/iReflect/reflect-app/apps/timetracker/serializers"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
	"github.com/iReflect/reflect-app/workers"
	"time"
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
		Preload("SprintMembers.Tasks").
		Find(&sprint).Error; err != nil {
		return err
	}
	tx := db.Begin()

	for _, sprintMember := range sprint.SprintMembers {
		for _, sprintMemberTask := range sprintMember.Tasks {
			if err := tx.Delete(&sprintMemberTask).Error; err != nil {
				tx.Rollback()
				return errors.New("sprint couldn't be deleted")
			}
		}
		if err := tx.Delete(&sprintMember).Error; err != nil {
			tx.Rollback()
			return errors.New("sprint couldn't be deleted")
		}
	}
	sprint.Status = retroModels.DeletedSprint
	if err := tx.Save(&sprint).Error; err != nil {
		tx.Rollback()
		return errors.New("sprint couldn't be deleted")
	}
	return tx.Commit().Error
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

// FreezeSprint freezes the given sprint
func (service SprintService) FreezeSprint(sprintID string) error {
	db := service.DB
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Where("status = ?", retroModels.ActiveSprint).
		Find(&sprint).Error; err != nil {
		return err
	}

	sprint.Status = retroModels.CompletedSprint
	if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
		return errors.New("sprint couldn't be frozen")
	}
	return nil
}

// Get return details of the given sprint
func (service SprintService) Get(sprintID string, userID uint) (*retrospectiveSerializers.Sprint, error) {
	db := service.DB
	var sprint retrospectiveSerializers.Sprint
	if err := db.Model(&retroModels.Sprint{}).
		Where("id = ?", sprintID).
		Preload("CreatedBy").
		Find(&sprint).Error; err != nil {
		return nil, err
	}
	return &sprint, nil
}

// AddSprintMember ...
func (service SprintService) AddSprintMember(sprintID string, memberID uint) (*retrospectiveSerializers.SprintMemberSummary, error) {
	db := service.DB
	var sprintMember retroModels.SprintMember
	sprintMemberSummary := new(retrospectiveSerializers.SprintMemberSummary)
	var sprint retroModels.Sprint

	err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprintID).
		Where("member_id = ?", memberID).
		Find(&retroModels.SprintMember{}).
		Error

	if err == nil {
		return nil, errors.New("Member already a part of the sprint")
	}

	err = db.Model(&retroModels.Sprint{}).
		Joins("JOIN retrospectives ON retrospectives.id=sprints.retrospective_id").
		Joins("JOIN user_teams ON retrospectives.team_id=user_teams.team_id").
		Where("user_teams.user_id=?", memberID).
		Where("sprints.id=?", sprintID).
		Find(&sprint).
		Error
	if err != nil {
		return nil, errors.New("Member is not a part of the retrospective team")
	}

	intSprintID, err := strconv.Atoi(sprintID)
	if err != nil {
		return nil, err
	}

	sprintMember.SprintID = uint(intSprintID)
	sprintMember.MemberID = memberID
	sprintMember.Vacations = 0
	sprintMember.Rating = retrospective.OkayRating
	sprintMember.AllocationPercent = 100
	sprintMember.ExpectationPercent = 100

	err = db.Create(&sprintMember).Error
	if err != nil {
		return nil, err
	}
	workers.Enqueuer.EnqueueUnique("sync_sprint_member_data", work.Q{"sprintMemberID": sprintMember.ID})

	sprintWorkingDays := utils.GetWorkingDaysBetweenTwoDates(*sprint.StartDate, *sprint.EndDate, true)
	if err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprint.ID).
		Joins("LEFT JOIN users ON users.id = sprint_members.member_id").
		Select("DISTINCT sprint_members.*, users.*").
		Scan(&sprintMemberSummary).
		Error; err != nil {
		return nil, err
	}

	sprintMemberSummary.ActualVelocity = 0
	memberWorkingDays := float64(sprintWorkingDays) - sprintMemberSummary.Vacations
	sprintMemberSummary.ExpectedVelocity = math.Floor((memberWorkingDays * 8.00 / sprint.Retrospective.HrsPerStoryPoint) *
		(sprintMemberSummary.ExpectationPercent / 100.00) * (sprintMemberSummary.AllocationPercent / 100.00))

	return sprintMemberSummary, nil
}

// RemoveSprintMember ...
func (service SprintService) RemoveSprintMember(sprintID string, memberID string) error {
	db := service.DB
	var sprintMember retroModels.SprintMember

	err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprintID).
		Where("id = ?", memberID).
		Preload("Tasks").
		Find(&sprintMember).
		Error

	if err != nil {
		return errors.New("Member not a part of the sprint")
	}

	tx := db.Begin()
	for _, smt := range sprintMember.Tasks {
		err = tx.Delete(&smt).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Delete(&sprintMember).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
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

	sprint.CurrentlySyncing = true
	err = db.Save(&sprint).Error
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
		err = service.SyncSprintMemberData(strconv.Itoa(int(sprintMember.ID)), false)
		if err != nil {
			return err
		}
	}

	// ToDo: Store tickets not in SMT
	// Maybe a Join table ST

	var currentTime time.Time
	currentTime = time.Now()
	sprint.LastSyncedAt = &currentTime
	sprint.CurrentlySyncing = false
	err = db.Save(&sprint).Error
	if err != nil {
		return err
	}

	return nil
}

// SyncSprintMemberData ...
func (service SprintService) SyncSprintMemberData(sprintMemberID string, independentRun bool) (err error) {
	db := service.DB
	var sprintMember retroModels.SprintMember
	err = db.Model(&retroModels.SprintMember{}).
		Where("id = ?", sprintMemberID).
		Preload("Sprint").
		Preload("Sprint.Retrospective").
		Find(&sprintMember).Error

	if err != nil {
		return err
	}

	sprint := sprintMember.Sprint

	if independentRun {
		sprint.CurrentlySyncing = true
		err = db.Save(&sprint).Error
		if err != nil {
			return err
		}
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

	if independentRun {
		var currentTime time.Time
		currentTime = time.Now()
		sprint.LastSyncedAt = &currentTime
		sprint.CurrentlySyncing = false
		err = db.Save(&sprint).Error
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

// GetSprintMembersSummary returns the sprint member summary list
func (service SprintService) GetSprintMembersSummary(sprintID string) (sprintMemberSummaryList *retrospectiveSerializers.SprintMemberSummaryListSerializer, err error) {
	db := service.DB
	sprintMemberSummaryList = new(retrospectiveSerializers.SprintMemberSummaryListSerializer)

	var sprint retroModels.Sprint
	if err = db.Where("id = ?", sprintID).
		Preload("Retrospective").
		Find(&sprint).
		Error; err != nil {
		return nil, err
	}
	sprintWorkingDays := utils.GetWorkingDaysBetweenTwoDates(*sprint.StartDate, *sprint.EndDate, true)
	if err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprint.ID).
		Joins("JOIN users ON users.id = sprint_members.member_id").
		Joins("LEFT JOIN sprint_member_tasks AS smt ON smt.sprint_member_id = sprint_members.id").
		Select("DISTINCT sprint_members.*, users.*, " +
			"SUM(smt.points_earned) over (PARTITION BY sprint_members.id) as actual_velocity").
		Scan(&sprintMemberSummaryList.Members).
		Error; err != nil {
		return nil, err
	}
	for _, sprintMemberSummary := range sprintMemberSummaryList.Members {
		memberWorkingDays := float64(sprintWorkingDays) - sprintMemberSummary.Vacations
		sprintMemberSummary.ExpectedVelocity = math.Floor((memberWorkingDays * 8.00 / sprint.Retrospective.HrsPerStoryPoint) *
			(sprintMemberSummary.ExpectationPercent / 100.00) * (sprintMemberSummary.AllocationPercent / 100.00))

	}
	return sprintMemberSummaryList, nil
}

// GetSprintMemberList returns the sprint member list
func (service SprintService) GetSprintMemberList(sprintID string) (sprintMemberList *userSerializers.MembersSerializer, err error) {
	db := service.DB
	sprintMemberList = new(userSerializers.MembersSerializer)

	if err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprintID).
		Joins("JOIN users ON users.id = sprint_members.member_id").
		Select("sprint_members.id, users.email, users.first_name, users.last_name, users.active").
		Scan(&sprintMemberList.Members).
		Error; err != nil {
		return nil, err
	}
	return sprintMemberList, nil
}

// UpdateSprintMember updates the sprint member summary
func (service SprintService) UpdateSprintMember(sprintID string, sprintMemberID string, memberData retrospectiveSerializers.SprintMemberSummary) (*retrospectiveSerializers.SprintMemberSummary, error) {
	db := service.DB

	var sprintMember retroModels.SprintMember
	if err := db.Model(&retroModels.SprintMember{}).
		Where("id = ?", sprintMemberID).
		Where("sprint_id = ?", sprintID).
		Preload("Sprint.Retrospective").
		Find(&sprintMember).
		Error; err != nil {
		return nil, err
	}

	sprintMember.AllocationPercent = memberData.AllocationPercent
	sprintMember.ExpectationPercent = memberData.ExpectationPercent
	sprintMember.Vacations = memberData.Vacations
	sprintMember.Rating = retrospective.Rating(memberData.Rating)
	sprintMember.Comment = memberData.Comment

	if err := db.Save(&sprintMember).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.id = ?", sprintMemberID).
		Joins("LEFT JOIN sprint_member_tasks AS smt ON smt.sprint_member_id = sprint_members.id").
		Select("SUM(smt.points_earned)").
		Group("sprint_members.id").
		Row().
		Scan(&memberData.ActualVelocity); err != nil {
		return nil, err
	}

	sprintWorkingDays := utils.GetWorkingDaysBetweenTwoDates(*sprintMember.Sprint.StartDate, *sprintMember.Sprint.EndDate, true)
	memberWorkingDays := float64(sprintWorkingDays - int(sprintMember.Vacations))

	memberData.ExpectedVelocity = math.Floor((memberWorkingDays * 8.00 / sprintMember.Sprint.Retrospective.HrsPerStoryPoint) *
		(float64(sprintMember.ExpectationPercent) / 100.00) * (float64(sprintMember.AllocationPercent) / 100.00))

	return &memberData, nil
}
