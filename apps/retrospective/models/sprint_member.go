package models

import (
	"errors"

	"github.com/jinzhu/gorm"

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
