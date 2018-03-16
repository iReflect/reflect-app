package models

import (
	"errors"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/sirupsen/logrus"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/retrospective"
)

// MemberTaskRoleValues ...
var MemberTaskRoleValues = [...]string{
	"Implementor",
	"Reviewer",
	"Validator",
}

// MemberTaskRole ...
type MemberTaskRole int8

// GetStringValue ...
func (role MemberTaskRole) GetStringValue() string {
	return MemberTaskRoleValues[role]
}

// MemberTaskRole
const (
	Implementor MemberTaskRole = iota
	Reviewer
	Validator
)

// SprintMemberTask represents a task for a member for a particular sprint
type SprintMemberTask struct {
	gorm.Model
	SprintMember     SprintMember
	SprintMemberID   uint `gorm:"not null"`
	Task             Task
	TaskID           uint                 `gorm:"not null"`
	TimeSpentMinutes uint                 `gorm:"not null"`
	PointsEarned     float64              `gorm:"default:0; not null"`
	PointsAssigned   float64              `gorm:"default:0; not null"`
	Rating           retrospective.Rating `gorm:"default:2; not null"`
	Comment          string               `gorm:"type:text"`
	Role             MemberTaskRole       `gorm:"default:0; not null"`
}

// BeforeSave ...
func (sprintMemberTask *SprintMemberTask) BeforeSave(db *gorm.DB) (err error) {
	var pointSum float64
	var task Task
	// TaskID is set when we use gorm and Task.ID is set when we use QOR admin,
	// so we need to add checks for both the cases.
	if sprintMemberTask.Task.ID == 0 {
		if err = db.Where("id = ?", sprintMemberTask.TaskID).Find(&task).Error; err != nil {
			return err
		}
	} else {
		task = sprintMemberTask.Task
	}
	db.Model(SprintMemberTask{}).Where("task_id = ? AND id <> ?", task.ID, sprintMemberTask.ID).Select("SUM(points_earned)").Row().Scan(&pointSum)

	// Sum of points earned for a task across all sprintMembers should not exceed the task's estimate
	if task.Estimate != nil && pointSum+sprintMemberTask.PointsEarned > *task.Estimate {
		err = errors.New("cannot earn more than estimate")
		return err
	}

	return
}

// RegisterSprintMemberTaskToAdmin ...
func RegisterSprintMemberTaskToAdmin(Admin *admin.Admin, config admin.Config) {
	sprintMemberTask := Admin.AddResource(&SprintMemberTask{}, &config)
	roleMeta := getMemberTaskRoleFieldMeta()
	sprintMemberTask.Meta(&roleMeta)
}

// getMemberTaskRoleFieldMeta is the meta config for the role field
func getMemberTaskRoleFieldMeta() admin.Meta {
	return admin.Meta{
		Name: "Role",
		Type: "select_one",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			sprintMemberTask := value.(*SprintMemberTask)
			return strconv.Itoa(int(sprintMemberTask.Role))
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			sprintMemberTask := resource.(*SprintMemberTask)
			value, err := strconv.Atoi(metaValue.Value.([]string)[0])
			if err != nil {
				logrus.Error("Cannot convert string to int")
				return
			}
			sprintMemberTask.Role = MemberTaskRole(value)
		},
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			for index, value := range MemberTaskRoleValues {
				results = append(results, []string{strconv.Itoa(index), value})
			}
			return
		},
		FormattedValuer: func(value interface{}, context *qor.Context) interface{} {
			sprintMemberTask := value.(*SprintMemberTask)
			return sprintMemberTask.Role.GetStringValue()
		},
	}
}
