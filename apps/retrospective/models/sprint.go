package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
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

// String ...
func (status SprintStatus) String() string {
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
	Title            string `gorm:"type:varchar(255); not null"`
	SprintID         string `gorm:"type:varchar(30); not null"`
	Retrospective    Retrospective
	RetrospectiveID  uint         `gorm:"not null"`
	Status           SprintStatus `gorm:"default:0; not null"`
	StartDate        *time.Time
	EndDate          *time.Time
	SprintMembers    []SprintMember
	LastSyncedAt     *time.Time
	CurrentlySyncing bool `gorm:"default:true;not null"`
	CreatedBy        userModels.User
	CreatedByID      uint `gorm:"not null"`
}

// BeforeSave ...
func (sprint *Sprint) BeforeSave(db *gorm.DB) (err error) {
	// end date should be after start date
	if sprint.StartDate != nil && sprint.EndDate != nil && sprint.EndDate.Before(*sprint.StartDate) {
		err = errors.New("end date can not be before start date")
		return err
	}

	if sprint.Status == ActiveSprint {
		sprints := []Sprint{}
		activeSprintCount := uint(0)

		db.LogMode(true)
		baseQuery := db.Where("retrospective_id = ?", sprint.RetrospectiveID)

		// More than one entries with status active for given retro should not be allowed
		baseQuery.Where("status = ? AND id <> ?", ActiveSprint, sprint.ID).Find(&sprints).Count(&activeSprintCount)
		if activeSprintCount > 0 {
			err = errors.New("another sprint is currently active")
			return err
		}

		// Active sprint must begin exactly 1 day after last completed sprint
		lastSprint := Sprint{}
		if !baseQuery.Where("status = ?", CompletedSprint).Order("end_date desc").First(&lastSprint).RecordNotFound() {
			expectedDate := lastSprint.EndDate.AddDate(0, 0, 1)
			if expectedDate.Year() != sprint.StartDate.Year() || expectedDate.YearDay() != sprint.StartDate.YearDay() {
				err = errors.New("sprint must begin the day after the last completed sprint ended")
				return err
			}
		}
	}

	return
}
