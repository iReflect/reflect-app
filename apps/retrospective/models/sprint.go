package models

import (
	"errors"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/sirupsen/logrus"
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

	if sprint.Status == ActiveSprint {
		sprints := []Sprint{}

		// RetrospectiveID is set when we use gorm and Retrospective.ID is set when we use QOR admin,
		// so we need to add checks for both the cases.
		retroID := sprint.RetrospectiveID
		if retroID == 0 {
			retroID = sprint.Retrospective.ID
		}
		baseQuery := db.Model(Sprint{}).Where("retrospective_id = ?", retroID)

		// More than one entries with status active for given retro should not be allowed
		baseQuery.Where("status = ? AND id <> ?", ActiveSprint, sprint.ID).Find(&sprints)
		if len(sprints) > 0 {
			err = errors.New("another sprint is currently active")
			return err
		}

		// Active sprint must begin exactly 1 day after last completed sprint
		lastSprint := Sprint{}
		if err := baseQuery.Where("status = ?", CompletedSprint).Order("end_date desc").First(&lastSprint).Error; err == nil {
			expectedDate := lastSprint.EndDate.AddDate(0, 0, 1)
			if expectedDate.Year() != sprint.StartDate.Year() || expectedDate.YearDay() != sprint.StartDate.YearDay() {
				err = errors.New("sprint must begin the day after the last completed sprint ended")
				return err
			}
		}
	}

	return
}

// BeforeSave ...
func (sprint *Sprint) BeforeSave(db *gorm.DB) (err error) {
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

	sprint.IndexAttrs("-SprintMembers", "-SyncStatus")
	sprint.NewAttrs("-SprintMembers", "-SyncStatus")
	sprint.EditAttrs("-SprintMembers", "-SyncStatus")
	sprint.ShowAttrs("-SprintMembers", "-SyncStatus")
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
	return db.Joins("JOIN sprint_members ON sprint_members.sprint_id = sprints.id").Where("sprint_members.deleted_at IS NULL")
}

// SprintJoinRetro ...
func SprintJoinRetro(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN retrospectives ON sprints.retrospective_id = retrospectives.id").Where("retrospectives.deleted_at IS NULL")
}
