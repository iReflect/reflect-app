package models

import (
	"errors"
	"strconv"

	"github.com/qor/qor"
	"github.com/qor/admin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/qor/qor/resource"

	"github.com/iReflect/reflect-app/libs/utils"
	"github.com/iReflect/reflect-app/apps/retrospective"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
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

// BeforeSave ...
func (sprintMember *SprintMember) BeforeSave(db *gorm.DB) (err error) {
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
		sprintWorkingDays := utils.GetWorkingDaysBetweenTwoDates(*sprint.StartDate, *sprint.EndDate, true)
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

// BeforeUpdate ...
func (sprintMember *SprintMember) BeforeUpdate(db *gorm.DB) (err error) {
	return sprintMember.BeforeSave(db)
}

// SMJoinSMT ...
func SMJoinSMT(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN sprint_member_tasks ON sprint_member_tasks.sprint_member_id = sprint_members.id").Where("sprint_member_tasks.deleted_at IS NULL")
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
