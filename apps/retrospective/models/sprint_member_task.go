package models

import (
	"errors"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/sirupsen/logrus"

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
	var sprintMember SprintMember
	// TaskID is set when we use gorm and Task.ID is set when we use QOR admin,
	// so we need to add checks for both the cases.
	if sprintMemberTask.Task.ID == 0 {
		if err = db.Where("id = ?", sprintMemberTask.TaskID).Find(&task).Error; err != nil {
			return err
		}
	} else {
		task = sprintMemberTask.Task
	}

	if sprintMemberTask.SprintMember.ID == 0 {
		if err = db.Where("id = ?", sprintMemberTask.SprintMemberID).Find(&sprintMember).Error; err != nil {
			return err
		}
	} else {
		sprintMember = sprintMemberTask.SprintMember
	}

	db.Model(SprintMemberTask{}).
		Where("task_id = ?", task.ID).
		Where("sprint_member_tasks.id <> ?", sprintMemberTask.ID).
		Joins("JOIN sprint_members AS sm ON sprint_member_tasks.sprint_member_id = sm.id").
		Joins("JOIN sprints ON sm.sprint_id = sprints.id").
		Where("(sprints.status <> ? OR sprints.id = ?)", DraftSprint, sprintMember.SprintID).
		Scopes(NotDeletedSprint).
		Select("SUM(points_earned)").Row().Scan(&pointSum)

	// Sum of points earned for a task across all sprintMembers should not exceed the task's estimate. Adding a 0.05 buffer for rounding errors
	// ToDo: Revisit to see if we can improve this.
	if pointSum+sprintMemberTask.PointsEarned > task.Estimate+0.05 {
		err = errors.New("cannot earn more than estimate")
		return err
	}

	return
}

// BeforeUpdate ...
func (sprintMemberTask *SprintMemberTask) BeforeUpdate(db *gorm.DB) (err error) {
	return sprintMemberTask.BeforeSave(db)
}

// RegisterSprintMemberTaskToAdmin ...
func RegisterSprintMemberTaskToAdmin(Admin *admin.Admin, config admin.Config) {
	sprintMemberTask := Admin.AddResource(&SprintMemberTask{}, &config)
	roleMeta := getMemberTaskRoleFieldMeta()
	taskMeta := getTaskMeta()
	sprintMembersMeta := getSprintMemberMeta()
	sprintMemberTask.Meta(&roleMeta)
	sprintMemberTask.Meta(&taskMeta)
	sprintMemberTask.Meta(&sprintMembersMeta)
}

// getSprintMemberMeta ...
func getSprintMemberMeta() admin.Meta {
	return admin.Meta{
		Name: "SprintMember",
		Type: "select_one",
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			db := context.GetDB()
			var members []SprintMember
			db.Model(&SprintMember{}).
				Preload("Member").
				Find(&members)

			for _, value := range members {
				results = append(results, []string{strconv.Itoa(int(value.ID)), "Sprint: " + strconv.Itoa(int(value.SprintID)) + " & Member: " + value.Member.FirstName + " " + value.Member.LastName})
			}
			return
		},
	}
}

// getTaskMeta ...
func getTaskMeta() admin.Meta {
	return admin.Meta{
		Name: "Task",
		Type: "select_one",
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			db := context.GetDB()
			var taskList []Task
			db.Model(&Task{}).Scan(&taskList)

			for _, value := range taskList {
				results = append(results, []string{strconv.Itoa(int(value.ID)), value.Key})
			}
			return
		},
	}
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

// SMTJoinTask ...
func SMTJoinTask(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN tasks ON sprint_member_tasks.task_id = tasks.id").Where("tasks.deleted_at IS NULL")
}
