package models

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/sirupsen/logrus"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/config"
	customErrors "github.com/iReflect/reflect-app/libs"
	"github.com/iReflect/reflect-app/libs/utils"
)

// SprintStatusValues ...
var SprintStatusValues = [...]string{
	"Draft",
	"Active",
	"Completed",
	"Deleted",
}

// SprintStatus ...
type SprintStatus int8

// GetStringValue ...
func (status SprintStatus) GetStringValue() string {
	return SprintStatusValues[status]
}

// SprintStatus
const (
	DraftSprint SprintStatus = iota
	ActiveSprint
	CompletedSprint
	DeletedSprint
)

// Sprint represents a sprint of a retrospective
type Sprint struct {
	gorm.Model
	Title           string `gorm:"type:varchar(255); not null"`
	SprintID        string `gorm:"type:varchar(30); not null"`
	Retrospective   Retrospective
	RetrospectiveID uint         `gorm:"not null"`
	Status          SprintStatus `gorm:"default:0; not null"`
	StartDate       *time.Time
	EndDate         *time.Time
	SprintMembers   []SprintMember
	SprintTasks     []SprintTask
	LastSyncedAt    *time.Time
	SyncStatus      []SprintSyncStatus
	CreatedBy       userModels.User
	CreatedByID     uint `gorm:"not null"`
}

// Validate ...
func (sprint *Sprint) Validate(db *gorm.DB) (err error) {
	// end date should be after start date
	if sprint.StartDate != nil && sprint.EndDate != nil && sprint.EndDate.Before(*sprint.StartDate) {
		err = errors.New("end date can not be before start date")
		return err
	}

	// RetrospectiveID is set when we use gorm and Retrospective.ID is set when we use QOR admin,
	// so we need to add checks for both the cases.
	retroID := sprint.RetrospectiveID
	if retroID == 0 {
		retroID = sprint.Retrospective.ID
	}
	baseQuery := db.Model(Sprint{}).
		Where("deleted_at IS NULL").
		Where("retrospective_id = ?", retroID).Scopes(NotDeletedSprint)

	if sprint.Status == DraftSprint {
		// More than one entries with status draft for given retro should not be allowed
		err = sprint.validateConcurrency(baseQuery, "another sprint is currently in draft")
		if err != nil {
			return err
		}
		// Draft sprint must begin exactly 1 day after last frozen/active sprint
		err = sprint.validateDateContinuity(baseQuery, []SprintStatus{CompletedSprint, ActiveSprint}, "sprint must begin the day after the last completed/activated sprint ended")
		if err != nil {
			return err
		}
	}

	if sprint.Status == ActiveSprint {
		// More than one entries with status active for given retro should not be allowed
		err = sprint.validateConcurrency(baseQuery, "another sprint is currently active")
		if err != nil {
			return err
		}
		// Active sprint must begin exactly 1 day after last completed sprint
		err = sprint.validateDateContinuity(baseQuery, []SprintStatus{CompletedSprint}, "sprint must begin the day after the last completed sprint ended")
		if err != nil {
			return err
		}
	}
	return
}

func (sprint *Sprint) validateConcurrency(baseQuery *gorm.DB, errorMessage string) (err error) {

	var sprints []Sprint

	baseQuery.Where("status = ? AND id <> ?", sprint.Status, sprint.ID).Find(&sprints)
	if len(sprints) > 0 {
		err = errors.New(errorMessage)
		return err
	}
	return
}

func (sprint *Sprint) validateDateContinuity(baseQuery *gorm.DB, statuses []SprintStatus, errorMessage string) (err error) {

	serverConf := config.GetConfig().Server
	location, err := time.LoadLocation(serverConf.TimeZone)

	if err != nil {
		log.Println("Invalid Timezone: ", err)
		utils.LogToSentry(err)
	}

	lastSprint := Sprint{}
	if err := baseQuery.Where("status IN (?)", statuses).
		Order("end_date desc").First(&lastSprint).Error; err == nil {
		expectedDate := lastSprint.EndDate.AddDate(0, 0, 1)
		startDate := *(sprint.StartDate)
		if location != nil {
			expectedDate = expectedDate.In(location)
			startDate = startDate.In(location)
		}
		if expectedDate.Year() != startDate.Year() || expectedDate.YearDay() != startDate.YearDay() {
			return &customErrors.ModelError{Message: errorMessage}
		}
	}
	return
}

// BeforeSave ...
func (sprint *Sprint) BeforeSave(db *gorm.DB) (err error) {
	sprint.SprintID = strings.TrimSpace(sprint.SprintID)
	sprint.Title = strings.TrimSpace(sprint.Title)
	return sprint.Validate(db)
}

// BeforeUpdate ...
func (sprint *Sprint) BeforeUpdate(db *gorm.DB) (err error) {
	return sprint.Validate(db)
}

// RegisterSprintToAdmin ...
func RegisterSprintToAdmin(Admin *admin.Admin, config admin.Config) {
	sprint := Admin.AddResource(&Sprint{}, &config)
	statusMeta := getSprintStatusFieldMeta()
	createdByMeta := userModels.GetUserFieldMeta("CreatedBy")

	sprint.Meta(&statusMeta)
	sprint.Meta(&createdByMeta)

	sprint.IndexAttrs("-SprintTasks", "-SprintMembers", "-SyncStatus")
	sprint.NewAttrs("-SprintTasks", "-SprintMembers", "-SyncStatus")
	sprint.EditAttrs("-SprintTasks", "-SprintMembers", "-SyncStatus")
	sprint.ShowAttrs("-SprintTasks", "-SprintMembers", "-SyncStatus")
}

// getSprintStatusFieldMeta is the meta config for the sprint status field
func getSprintStatusFieldMeta() admin.Meta {
	return admin.Meta{
		Name: "Status",
		Type: "select_one",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			sprint := value.(*Sprint)
			return strconv.Itoa(int(sprint.Status))
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			sprint := resource.(*Sprint)
			value, err := strconv.Atoi(metaValue.Value.([]string)[0])
			if err != nil {
				logrus.Error("Cannot convert string to int")
				return
			}
			sprint.Status = SprintStatus(value)
		},
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			for index, value := range SprintStatusValues {
				results = append(results, []string{strconv.Itoa(index), value})
			}
			return
		},
		FormattedValuer: func(value interface{}, context *qor.Context) interface{} {
			sprint := value.(*Sprint)
			return sprint.Status.GetStringValue()
		},
	}
}

// NotDeletedSprint is a gorm scope used to exclude the deleted/discard sprints
func NotDeletedSprint(db *gorm.DB) *gorm.DB {
	return db.Not("sprints.status = ?", DeletedSprint)
}

// SprintJoinSM ...
func SprintJoinSM(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprint_members ON sprint_members.sprint_id = sprints.id AND sprint_members.deleted_at IS NULL")
}

// SprintJoinRetro ...
func SprintJoinRetro(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN retrospectives ON sprints.retrospective_id = retrospectives.id AND retrospectives.deleted_at IS NULL")
}

// SprintJoinST ...
func SprintJoinST(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprint_tasks ON sprints.id = sprint_tasks.sprint_id AND sprint_tasks.deleted_at IS NULL")
}
