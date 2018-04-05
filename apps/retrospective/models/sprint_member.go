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
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/libs/utils"
)

// SprintMember represents a member of a particular sprint
type SprintMember struct {
	gorm.Model
	Sprint             Sprint
	SprintID           uint `gorm:"not null"`
	Member             userModels.User
	MemberID           uint    `gorm:"not null"`
	AllocationPercent  float64 `gorm:"not null;default:100"`
	ExpectationPercent float64 `gorm:"not null;default:100"`
	Tasks              []SprintMemberTask
	Vacations          float64              `gorm:"not null;default:0"`
	Rating             retrospective.Rating `gorm:"default:2; not null"`
	Comment            string               `gorm:"type:text"`
}

// Validate ...
func (sprintMember *SprintMember) Validate(db *gorm.DB) (err error) {
	var sprint Sprint
	if sprintMember.Sprint.ID == 0 {
		if err = db.Where("id = ?", sprintMember.SprintID).Find(&sprint).Error; err != nil {
			return errors.New("cannot find sprint")
		}
	} else {
		sprint = sprintMember.Sprint
	}
	// Vacations should not be longer than sprint duration
	if sprint.StartDate != nil && sprint.EndDate != nil {
		sprintWorkingDays := utils.GetWorkingDaysBetweenTwoDates(*sprint.StartDate, *sprint.EndDate)
		if sprintMember.Vacations > float64(sprintWorkingDays) {
			err = errors.New("vacations cannot be longer than sprint duration")
			return err
		}
	}
	if sprintMember.Vacations < 0 {
		err = errors.New("vacations cannot be negative")
		return err
	}
	if sprintMember.AllocationPercent < 0 {
		err = errors.New("allocation cannot be negative")
		return err
	}
	if sprintMember.ExpectationPercent < 0 {
		err = errors.New("expectation cannot be negative")
		return err
	}
	return
}

// BeforeSave ...
func (sprintMember *SprintMember) BeforeSave(db *gorm.DB) (err error) {
	return sprintMember.Validate(db)
}

// BeforeUpdate ...
func (sprintMember *SprintMember) BeforeUpdate(db *gorm.DB) (err error) {
	return sprintMember.Validate(db)
}

// RegisterSprintMemberToAdmin ...
func RegisterSprintMemberToAdmin(Admin *admin.Admin, config admin.Config) {
	sprintMember := Admin.AddResource(&SprintMember{}, &config)

	ratingMeta := getSprintMemberRatingMeta()
	memberMeta := userModels.GetUserFieldMeta("Member")

	sprintMember.Meta(&ratingMeta)
	sprintMember.Meta(&memberMeta)

	sprintMember.IndexAttrs("-Tasks")
	sprintMember.NewAttrs("-Tasks")
	sprintMember.EditAttrs("-Tasks")
	sprintMember.ShowAttrs("-Tasks")
}

// getSprintMemberRatingMeta ...
func getSprintMemberRatingMeta() admin.Meta {
	return admin.Meta{
		Name: "Rating",
		Type: "select_one",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			sprintMember := value.(*SprintMember)
			return strconv.Itoa(int(sprintMember.Rating))
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			sprintMember := resource.(*SprintMember)
			value, err := strconv.Atoi(metaValue.Value.([]string)[0])
			if err != nil {
				logrus.Error("Cannot convert string to int")
				return
			}
			sprintMember.Rating = retrospective.Rating(value)
		},
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			for index, value := range retrospective.RatingValues {
				results = append(results, []string{strconv.Itoa(index), value})
			}
			return
		},
		FormattedValuer: func(value interface{}, context *qor.Context) interface{} {
			sprintMember := value.(*SprintMember)
			return sprintMember.Rating.GetStringValue()
		},
	}
}

// SMJoinSMT ...
func SMJoinSMT(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprint_member_tasks ON sprint_member_tasks.sprint_member_id = sprint_members.id").Where("sprint_member_tasks.deleted_at IS NULL")
}

// SMLeftJoinSMT ...
func SMLeftJoinSMT(db *gorm.DB) *gorm.DB {
	return db.Joins("LEFT JOIN sprint_member_tasks ON sprint_member_tasks.sprint_member_id = sprint_members.id").Where("sprint_member_tasks.deleted_at IS NULL")
}

// SMJoinSprint ...
func SMJoinSprint(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprints ON sprint_members.sprint_id = sprints.id").Where("sprints.deleted_at IS NULL")
}

// SMJoinMember ...
func SMJoinMember(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN users ON sprint_members.member_id = users.id").Where("users.deleted_at IS NULL")
}
