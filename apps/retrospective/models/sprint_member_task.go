package models

import (
	"errors"
	"fmt"
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
	SprintTask       SprintTask
	SprintTaskID     uint                 `gorm:"not null"`
	TimeSpentMinutes uint                 `gorm:"not null"`
	PointsEarned     float64              `gorm:"default:0; not null"`
	PointsAssigned   float64              `gorm:"default:0; not null"`
	Rating           retrospective.Rating `gorm:"default:2; not null"`
	Comment          string               `gorm:"type:text"`
	Role             MemberTaskRole       `gorm:"default:0; not null"`
}

// Validate ...
func (sprintMemberTask *SprintMemberTask) Validate(db *gorm.DB) (err error) {
	var isInvalidPointSum bool
	sprintTaskFilter := db.Model(&SprintTask{}).Where("id = ?", sprintMemberTask.SprintTaskID).
		Select("task_id")
	sprintFilter := db.Model(&SprintMember{}).Where("id = ?", sprintMemberTask.SprintMemberID).
		Select("sprint_id").QueryExpr()

	// Sum of points earned for a task across all sprintMembers should not exceed the task's estimate.
	// Adding a 0.05 buffer for rounding errors
	// ToDo: Revisit to see if we can improve this.
	db.Model(SprintMemberTask{}).
		Where("sprint_member_tasks.id <> ?", sprintMemberTask.ID).
		Where("sprint_task.task_id = (?)", sprintTaskFilter).
		Scopes(SMTJoinST, STJoinTask, SMTJoinSM, SMJoinSprint).
		Where("(sprints.status <> ? OR sprints.id = (?))", DraftSprint, sprintFilter).
		Scopes(NotDeletedSprint).
		Select("(SUM(points_earned) > ( tasks.Estimate + 0.05))").Row().Scan(&isInvalidPointSum)

	if isInvalidPointSum {
		return errors.New("cannot earn more than estimate")
	}

	return
}

// BeforeSave ...
func (sprintMemberTask *SprintMemberTask) BeforeSave(db *gorm.DB) (err error) {
	return sprintMemberTask.Validate(db)
}

// BeforeUpdate ...
func (sprintMemberTask *SprintMemberTask) BeforeUpdate(db *gorm.DB) (err error) {
	return sprintMemberTask.Validate(db)
}

// RegisterSprintMemberTaskToAdmin ...
func RegisterSprintMemberTaskToAdmin(Admin *admin.Admin, config admin.Config) {
	sprintMemberTask := Admin.AddResource(&SprintMemberTask{}, &config)

	sprintTaskMeta := getSprintTaskMeta()
	roleMeta := getMemberTaskRoleFieldMeta()
	sprintMembersMeta := getSprintMemberMeta()
	ratingMeta := getSprintMemberTaskRatingMeta()

	sprintMemberTask.Meta(&sprintTaskMeta)
	sprintMemberTask.Meta(&roleMeta)
	sprintMemberTask.Meta(&ratingMeta)
	sprintMemberTask.Meta(&sprintMembersMeta)
}

// getSprintMemberTaskRatingMeta ...
func getSprintMemberTaskRatingMeta() admin.Meta {
	return admin.Meta{
		Name: "Rating",
		Type: "select_one",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			sprintMemberTask := value.(*SprintMemberTask)
			return strconv.Itoa(int(sprintMemberTask.Rating))
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			sprintMemberTask := resource.(*SprintMemberTask)
			value, err := strconv.Atoi(metaValue.Value.([]string)[0])
			if err != nil {
				logrus.Error("Cannot convert string to int")
				return
			}
			sprintMemberTask.Rating = retrospective.Rating(value)
		},
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			for index, value := range retrospective.RatingValues {
				results = append(results, []string{strconv.Itoa(index), value})
			}
			return
		},
		FormattedValuer: func(value interface{}, context *qor.Context) interface{} {
			sprintMemberTask := value.(*SprintMemberTask)
			return sprintMemberTask.Rating.GetStringValue()
		},
	}
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
				results = append(results, []string{
					strconv.Itoa(int(value.ID)),
					fmt.Sprintf("Sprint: %s & Member: %s %s",
						strconv.Itoa(int(value.SprintID)),
						value.Member.FirstName,
						value.Member.LastName)})
			}
			return
		},
	}
}

// getSprintTaskMeta ...
func getSprintTaskMeta() admin.Meta {
	return admin.Meta{
		Name: "SprintTask",
		Type: "select_one",
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			db := context.GetDB()
			var sprintTaskList []SprintTask
			db.Model(&SprintTask{}).Preload("Task").Scan(&sprintTaskList)

			for _, value := range sprintTaskList {
				results = append(results, []string{strconv.Itoa(int(value.ID)), value.Task.Key})
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

// SMTJoinST ...
func SMTJoinST(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprint_tasks ON sprint_member_tasks.sprint_task_id = sprint_tasks.id").
		Where("sprint_tasks.deleted_at IS NULL")
}

// SMTJoinSM ...
func SMTJoinSM(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprint_members ON sprint_member_tasks.sprint_member_id = sprint_members.id").
		Where("sprint_members.deleted_at IS NULL")
}
