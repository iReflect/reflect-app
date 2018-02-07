package models

import (
	"errors"

	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/retrospective"
)

// SprintMemberTask represents a task for a member for a particular sprint
type SprintMemberTask struct {
	gorm.Model
	SprintMember     SprintMember
	SprintMemberID   uint `gorm:"not null"`
	Task             Task
	TaskID           uint                 `gorm:"not null"`
	TimeSpentMinutes uint                 `gorm:"not null"`
	PointsEarned     float64              `gorm:"not null"`
	PointsAssigned   float64              `gorm:"not null"`
	Rating           retrospective.Rating `gorm:"default:0; not null"`
	Comment          string               `gorm:"type:text; not null"`
}

// BeforeSave ...
func (sprintMemberTask *SprintMemberTask) BeforeSave(db *gorm.DB) (err error) {
	var pointSum float64
	db.LogMode(true)
	db.Model(SprintMemberTask{}).Where("task_id = ? AND id <> ?", sprintMemberTask.TaskID, sprintMemberTask.ID).Select("SUM(points_earned)").Row().Scan(&pointSum)

	// Sum of points earned for a task across all sprintMembers should not exceed the task's estimate
	if pointSum+sprintMemberTask.PointsEarned > sprintMemberTask.Task.Estimate {
		err = errors.New("cannot earn more than estimate")
		return err
	}

	return
}
