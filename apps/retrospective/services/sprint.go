package services

import (
	"errors"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/gocraft/work"
	"github.com/iReflect/reflect-app/apps/retrospective"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	taskTrackerSerializers "github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"github.com/iReflect/reflect-app/apps/timetracker"
	timeTrackerSerializers "github.com/iReflect/reflect-app/apps/timetracker/serializers"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
	"github.com/iReflect/reflect-app/constants"
	"github.com/iReflect/reflect-app/libs/utils"
	"github.com/iReflect/reflect-app/workers"
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
		return errors.New(constants.SprintNotFound)
	}
	tx := db.Begin()

	for _, sprintMember := range sprint.SprintMembers {
		for _, sprintMemberTask := range sprintMember.Tasks {
			if err := tx.Delete(&sprintMemberTask).Error; err != nil {
				tx.Rollback()
				return errors.New(constants.SprintMemberTaskDeleteError)
			}
		}
		if err := tx.Delete(&sprintMember).Error; err != nil {
			tx.Rollback()
			return errors.New(constants.SprintMemberDeleteError)
		}
	}
	sprint.Status = retroModels.DeletedSprint
	if err := tx.Save(&sprint).Error; err != nil {
		tx.Rollback()
		return errors.New(constants.SprintDeleteError)
	}
	if err := tx.Commit().Error; err != nil {
		return errors.New(constants.SprintDeleteError)
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
		return errors.New(constants.SprintNotFound)
	}

	sprint.Status = retroModels.ActiveSprint
	if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
		return errors.New(constants.SprintActivateError)
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
		return errors.New(constants.SprintNotFound)
	}

	// ToDo: Check SM count > 0

	sprint.Status = retroModels.CompletedSprint
	if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
		return errors.New(constants.SprintFreezeError)
	}
	return nil
}

// Get return details of the given sprint
func (service SprintService) Get(sprintID string) (*retrospectiveSerializers.Sprint, error) {
	db := service.DB
	var sprint retrospectiveSerializers.Sprint
	if err := db.Model(&retroModels.Sprint{}).
		Where("id = ?", sprintID).
		Preload("CreatedBy").
		Find(&sprint).Error; err != nil {
		return nil, errors.New(constants.SprintNotFound)
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
		// It means a sprint member already exists
		return nil, errors.New(constants.AlreadySprintMemberError)
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
		return nil, errors.New("member is not a part of the retrospective team")
	}

	intSprintID, err := strconv.Atoi(sprintID)
	if err != nil {
		return nil, errors.New(constants.InvalidSprintID)
	}

	sprintMember.SprintID = uint(intSprintID)
	sprintMember.MemberID = memberID
	sprintMember.Vacations = 0
	sprintMember.Rating = retrospective.OkayRating
	sprintMember.AllocationPercent = 100
	sprintMember.ExpectationPercent = 100

	err = db.Create(&sprintMember).Error
	if err != nil {
		return nil, errors.New(constants.SprintMemberCreateError)
	}

	workers.Enqueuer.EnqueueUnique("sync_sprint_member_data", work.Q{"sprintMemberID": strconv.Itoa(int(sprintMember.ID))})

	if err = db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprint.ID).
		Joins("LEFT JOIN users ON users.id = sprint_members.member_id").
		Select("DISTINCT sprint_members.*, users.*").
		Scan(&sprintMemberSummary).
		Error; err != nil {
		return nil, errors.New(constants.SprintMemberSummaryError)
	}

	sprintMemberSummary.ActualStoryPoint = 0
	sprintMemberSummary.SetExpectedStoryPoint(sprint, sprint.Retrospective)

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
		return errors.New(constants.NotASprintMemberError)
	}

	tx := db.Begin()
	for _, smt := range sprintMember.Tasks {
		err = tx.Delete(&smt).Error
		if err != nil {
			tx.Rollback()
			return errors.New(constants.SprintMemberTaskDeleteError)
		}
	}

	err = tx.Delete(&sprintMember).Error
	if err != nil {
		tx.Rollback()
		return errors.New(constants.SprintMemberDeleteError)
	}

	if err := tx.Commit().Error; err != nil {
		return errors.New(constants.SprintMemberDeleteError)
	}
	return nil
}

// SyncSprintData ...
func (service SprintService) SyncSprintData(sprintID string) error {
	db := service.DB
	var sprint retroModels.Sprint
	err := db.Model(&retroModels.Sprint{}).
		Scopes(retroModels.NotDeletedSprint).
		Where("id = ?", sprintID).
		Preload("SprintMembers").
		Preload("Retrospective").
		Find(&sprint).Error

	if err != nil {
		utils.LogToSentry(err)
		return errors.New(constants.SprintNotFound)
	}

	sprint.CurrentlySyncing = true
	err = db.Save(&sprint).Error
	if err != nil {
		utils.LogToSentry(err)
		return errors.New(constants.SprintUpdateError)
	}

	taskProviderConfig, err := tasktracker.DecryptTaskProviders(sprint.Retrospective.TaskProviderConfig)
	if err != nil {
		utils.LogToSentry(err)
		return errors.New(constants.TaskProviderConfigParseError)
	}

	tickets, err := tasktracker.GetSprintTaskList(taskProviderConfig, sprint.SprintID)
	if err != nil {
		utils.LogToSentry(err)
		return errors.New(constants.SprintTaskListError)
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
			utils.LogToSentry(err)
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
		utils.LogToSentry(err)
		return errors.New(constants.SprintSyncError)
	}

	return nil
}

// SyncSprintMemberData ...
func (service SprintService) SyncSprintMemberData(sprintMemberID string, independentRun bool) error {
	db := service.DB
	var sprintMember retroModels.SprintMember
	err := db.Model(&retroModels.SprintMember{}).
		Where("id = ?", sprintMemberID).
		Preload("Sprint").
		Preload("Member").
		Preload("Sprint.Retrospective").
		Find(&sprintMember).Error

	if err != nil {
		utils.LogToSentry(err)
		return errors.New(constants.SprintMemberNotFound)
	}

	sprint := sprintMember.Sprint

	if independentRun {
		sprint.CurrentlySyncing = true
		err = db.Save(&sprint).Error
		if err != nil {
			utils.LogToSentry(err)
			return errors.New(constants.SprintUpdateError)
		}
	}

	if sprint.StartDate == nil || sprint.EndDate == nil {
		utils.LogToSentry(err)
		return errors.New("sprint has no start/end date")
	}

	timeLogs, err := timetracker.GetProjectTimeLogs(sprintMember.Member.TimeProviderConfig, sprint.Retrospective.ProjectName, *sprint.StartDate, *sprint.EndDate)

	if err != nil {
		utils.LogToSentry(err)
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
		return errors.New(constants.TaskProviderConfigParseError)
	}

	tickets, err := tasktracker.GetTaskList(taskProviderConfig, ticketIDs)
	if err != nil {
		utils.LogToSentry(err)
		return err
	}

	for _, ticket := range tickets {
		err = service.addOrUpdateTask(ticket, sprintMember.Sprint.Retrospective.ID)
		if err != nil {
			utils.LogToSentry(err)
			return err
		}
	}

	// Reset existing time_spent
	err = db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_id = ?", sprintMemberID).
		UpdateColumn("time_spent_minutes", 0).Error

	if err != nil {
		utils.LogToSentry(err)
		return errors.New(constants.SprintMemberTaskUpdateError)
	}

	for _, timeLog := range timeLogs {
		err = service.addOrUpdateSMT(timeLog, sprintMember.ID, sprint.RetrospectiveID)
		if err != nil {
			utils.LogToSentry(err)
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
			utils.LogToSentry(err)
			return errors.New(constants.SprintUpdateError)
		}
	}

	return nil
}

func (service SprintService) insertTask(ticketID string, retroID uint) (err error) {
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
	// ToDo: Handle moved issues! ie ticket id changes
	db := service.DB
	var task retroModels.Task
	err = db.Where(retroModels.Task{RetrospectiveID: retroID, TaskID: ticket.ID}).
		Assign(retroModels.Task{
			Summary:  ticket.Summary,
			Type:     ticket.Type,
			Priority: ticket.Priority,
			Estimate: ticket.Estimate,
			Assignee: ticket.Assignee,
			Status:   ticket.Status,
		}).
		FirstOrCreate(&task).Error

	if err != nil {
		utils.LogToSentry(err)
		return err
	}

	return nil
}

func (service SprintService) addOrUpdateSMT(timeLog timeTrackerSerializers.TimeLog, sprintMemberID uint, retroID uint) error {
	db := service.DB
	var sprintMemberTask retroModels.SprintMemberTask
	var task retroModels.Task
	err := db.Model(&retroModels.SprintMemberTask{}).
		Where("sprint_member_id = ?", sprintMemberID).
		Joins("JOIN tasks ON tasks.id=sprint_member_tasks.task_id").
		Where("tasks.task_id = ?", timeLog.TaskID).
		Where("tasks.retrospective_id = ?", retroID).
		FirstOrInit(&sprintMemberTask).Error
	if err != nil {
		return errors.New(constants.SprintMemberTaskNotFound)
	}

	err = db.Model(&retroModels.Task{}).
		Where("task_id = ?", timeLog.TaskID).
		Where("tasks.retrospective_id = ?", retroID).
		First(&task).Error
	if err != nil {
		return nil
	}

	sprintMemberTask.SprintMemberID = sprintMemberID
	sprintMemberTask.TaskID = task.ID
	sprintMemberTask.TimeSpentMinutes = timeLog.Minutes

	if err := db.Save(&sprintMemberTask).Error; err != nil {
		return errors.New(constants.SprintMemberTaskUpdateError)
	}
	return nil
}

// GetSprintsList ...
func (service SprintService) GetSprintsList(retrospectiveID string, userID uint) (*retrospectiveSerializers.SprintsSerializer, error) {
	db := service.DB
	sprints := new(retrospectiveSerializers.SprintsSerializer)

	err := db.Model(&retroModels.Sprint{}).
		Where("retrospective_id = ?", retrospectiveID).
		Where("status in (?) OR (status = (?) AND created_by_id = (?))", []retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}, retroModels.DraftSprint, userID).
		Preload("CreatedBy").
		Order("end_date desc").
		Find(&sprints.Sprints).Error

	if err != nil {
		return nil, errors.New(constants.SprintNotFound)
	}
	return sprints, nil
}

// GetSprintMembersSummary returns the sprint member summary list
func (service SprintService) GetSprintMembersSummary(sprintID string) (*retrospectiveSerializers.SprintMemberSummaryListSerializer, error) {
	db := service.DB
	sprintMemberSummaryList := new(retrospectiveSerializers.SprintMemberSummaryListSerializer)

	var sprint retroModels.Sprint
	if err := db.Where("id = ?", sprintID).
		Preload("Retrospective").
		Find(&sprint).
		Error; err != nil {
		return nil, errors.New(constants.SprintNotFound)
	}
	if err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprint.ID).
		Joins("JOIN users ON users.id = sprint_members.member_id").
		Joins("LEFT JOIN sprint_member_tasks AS smt ON smt.sprint_member_id = sprint_members.id").
		Select("DISTINCT sprint_members.*, users.*, " +
			"SUM(smt.points_earned) over (PARTITION BY sprint_members.id) as actual_story_point, " +
			"SUM(smt.time_spent_minutes) over (PARTITION BY sprint_members.id) " +
			"as total_time_spent_in_min").
		Scan(&sprintMemberSummaryList.Members).
		Error; err != nil {
		return nil, errors.New(constants.SprintMemberSummaryError)
	}
	for _, sprintMemberSummary := range sprintMemberSummaryList.Members {
		sprintMemberSummary.SetExpectedStoryPoint(sprint, sprint.Retrospective)
	}
	return sprintMemberSummaryList, nil
}

// GetSprintMemberList returns the sprint member list
func (service SprintService) GetSprintMemberList(sprintID string) (*userSerializers.MembersSerializer, error) {
	db := service.DB
	sprintMemberList := new(userSerializers.MembersSerializer)

	if err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_id = ?", sprintID).
		Joins("JOIN users ON users.id = sprint_members.member_id").
		Select("sprint_members.id, users.email, users.first_name, users.last_name, users.active").
		Scan(&sprintMemberList.Members).
		Error; err != nil {
		return nil, errors.New(constants.SprintMemberNotFound)
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
		return nil, errors.New(constants.SprintMemberNotFound)
	}

	sprintMember.AllocationPercent = memberData.AllocationPercent
	sprintMember.ExpectationPercent = memberData.ExpectationPercent
	sprintMember.Vacations = memberData.Vacations
	sprintMember.Rating = retrospective.Rating(memberData.Rating)
	sprintMember.Comment = memberData.Comment

	if err := db.Save(&sprintMember).Error; err != nil {
		return nil, errors.New(constants.SprintMemberUpdateError)
	}

	if err := db.Model(&retroModels.SprintMember{}).
		Where("sprint_members.id = ?", sprintMemberID).
		Joins("LEFT JOIN sprint_member_tasks AS smt ON smt.sprint_member_id = sprint_members.id").
		Select("COALESCE(SUM(smt.points_earned), 0) as actual_story_point, "+
			"COALESCE(SUM(smt.time_spent_minutes), 0) as total_time_spent_in_min").
		Group("sprint_members.id").
		Row().
		Scan(&memberData.ActualStoryPoint, &memberData.TotalTimeSpentInMin); err != nil {
		return nil, errors.New(constants.SprintMemberSummaryError)
	}

	memberData.SetExpectedStoryPoint(sprintMember.Sprint, sprintMember.Sprint.Retrospective)

	return &memberData, nil
}

// Create creates a new sprint for the retro
func (service SprintService) Create(retroID string, sprintData retrospectiveSerializers.CreateSprintSerializer) (*retroModels.Sprint, error) {
	db := service.DB
	var err error
	var sprint retroModels.Sprint
	var retro retroModels.Retrospective
	var sprintMember retroModels.SprintMember

	if err := db.Model(&retro).
		Where("id = ?", retroID).
		Find(&retro).Error; err != nil {
		return nil, errors.New(constants.RetrospectiveNotFound)
	}

	var teamMemberIDs []uint

	err = db.Model(&retroModels.Sprint{}).
		Joins("JOIN sprint_members ON sprint_members.sprint_id = sprints.id").
		Where("sprints.status in (?)", []retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}).
		Where("sprints.retrospective_id = ?", retro.ID).
		Order("sprints.end_date DESC, sprints.created_at DESC").
		Pluck("sprint_members.member_id", &teamMemberIDs).
		Error
	if err != nil || len(teamMemberIDs) < 1 {
		err = db.Model(&userModels.UserTeam{}).
			Where("team_id = ?", retro.TeamID).
			Where("leaved_at IS NULL OR leaved_at > NOW()").
			Pluck("DISTINCT user_id", &teamMemberIDs).
			Error
		if err != nil {
			return nil, errors.New(constants.SprintMemberAddError)
		}
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
			return nil, errors.New(constants.TaskProviderConfigParseError)
		}

		connections, err := tasktracker.GetConnections(taskProviderConfig)
		if err != nil {
			return nil, err
		}

		var providerSprint *taskTrackerSerializers.Sprint

		for _, connection := range connections {
			providerSprint = connection.GetSprint(sprintData.SprintID)
			if providerSprint != nil {
				if providerSprint.FromDate == nil || providerSprint.ToDate == nil {
					return nil, errors.New(constants.SprintStartEndDateError)
				}
				sprint.StartDate = providerSprint.FromDate
				sprint.EndDate = providerSprint.ToDate
				break
			}
		}

		if providerSprint == nil {
			return nil, errors.New(constants.SprintNotFoundInTaskTracker)
		}
	}

	tx := db.Begin() // transaction begin

	if err = tx.Create(&sprint).Error; err != nil {
		tx.Rollback()
		return nil, errors.New(constants.SprintCreateError)
	}

	for _, userID := range teamMemberIDs {
		sprintMember = retroModels.SprintMember{
			SprintID:  uint(sprint.ID),
			MemberID:  userID,
			Vacations: 0,
			Rating:    retrospective.OkayRating,
			// TODO: Instead of setting it to default to 100%,
			// we can use the previous active sprint's data for the allocation and expectation values
			AllocationPercent:  100,
			ExpectationPercent: 100,
		}
		if err = tx.Create(&sprintMember).Error; err != nil {
			tx.Rollback()
			return nil, errors.New(constants.SprintMemberCreateError)
		}
	}

	workers.Enqueuer.EnqueueUnique("sync_sprint_data", work.Q{"sprintID": strconv.Itoa(int(sprint.ID)), "assignPoints": true})
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New(constants.SprintCreateError)
	}
	return &sprint, nil
}

// UpdateSprint updates the given sprint
func (service SprintService) UpdateSprint(sprintID string, sprintData retrospectiveSerializers.UpdateSprintSerializer) (*retrospectiveSerializers.Sprint, error) {
	db := service.DB
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Find(&sprint).Error; err != nil {
		return nil, errors.New(constants.SprintNotFound)
	}

	if sprintData.GoodHighlights != nil {
		sprint.GoodHighlights = *sprintData.GoodHighlights
	}

	if sprintData.OkayHighlights != nil {
		sprint.OkayHighlights = *sprintData.OkayHighlights
	}

	if sprintData.BadHighlights != nil {
		sprint.BadHighlights = *sprintData.BadHighlights
	}

	if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
		return nil, errors.New(constants.SprintUpdateError)
	}

	return service.Get(sprintID)
}

// AssignPoints ...
func (service SprintService) AssignPoints(sprintID string) error {
	db := service.DB
	var sprint retroModels.Sprint
	err := db.Model(&retroModels.Sprint{}).
		Scopes(retroModels.NotDeletedSprint).
		Where("id = ?", sprintID).
		Preload("SprintMembers").
		Preload("Retrospective").
		Find(&sprint).Error

	if err != nil {
		utils.LogToSentry(err)
		return errors.New(constants.SprintNotFound)
	}

	sprint.CurrentlySyncing = true
	err = db.Save(&sprint).Error
	if err != nil {
		utils.LogToSentry(err)
		return errors.New(constants.SprintUpdateError)
	}

	dbs := db.Model(retroModels.SprintMemberTask{}).
		Joins("JOIN sprint_members AS sm ON sprint_member_tasks.sprint_member_id = sm.id").
		Joins("JOIN tasks ON tasks.id = sprint_member_tasks.task_id").
		Joins("JOIN sprints ON sm.sprint_id = sprints.id").
		Where("(sprints.status <> ? OR sprints.id = ?)", retroModels.DraftSprint, sprintID).
		Scopes(retroModels.NotDeletedSprint).
		Where("tasks.retrospective_id = ?", sprint.RetrospectiveID).
		Select("sprint_member_tasks.*," +
			"row_number() over (PARTITION BY sprint_member_tasks.task_id, sm.sprint_id order by sprint_member_tasks.time_spent_minutes desc) as time_spent_rank, " +
			"sm.sprint_id, " +
			"(tasks.estimate - (SUM(sprint_member_tasks.points_earned) over (PARTITION BY sprint_member_tasks.task_id))) as remaining_points").
		QueryExpr()

	err = db.Exec("UPDATE sprint_member_tasks "+
		"SET points_assigned = s1.remaining_points, points_earned = s1.remaining_points "+
		"FROM (?) AS s1 "+
		"WHERE s1.sprint_id = ? and time_spent_rank = 1 and sprint_member_tasks.id = s1.id;", dbs, sprintID).Error

	if err != nil {
		utils.LogToSentry(err)
	}
	sprint.CurrentlySyncing = false
	err = db.Save(&sprint).Error
	if err != nil {
		utils.LogToSentry(err)
		return errors.New(constants.SprintUpdateError)
	}

	return nil
}
