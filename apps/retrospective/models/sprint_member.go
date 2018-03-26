package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"

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

// BeforeSave ...
func (sprintMember *SprintMember) BeforeSave(db *gorm.DB) (err error) {
	//ToDo: Investigate failing association during SMT save
	// Vacations should not be more than sprint length
	if sprintMember.Sprint.StartDate != nil && sprintMember.Sprint.EndDate != nil {
		sprintWorkingDays := utils.GetWorkingDaysBetweenTwoDates(*sprintMember.Sprint.StartDate, *sprintMember.Sprint.EndDate, true)
		if sprintMember.Vacations > float64(sprintWorkingDays) {
			err = errors.New("vacations cannot be more than sprint length")
			return err
		}
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
	sprintMember.IndexAttrs("-Tasks")
	sprintMember.NewAttrs("-Tasks")
	sprintMember.EditAttrs("-Tasks")
	sprintMember.ShowAttrs("-Tasks")
}
