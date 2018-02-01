package models

import (
	"errors"

	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/retrospective"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

// SprintMember represents a member of a particular sprint
type SprintMember struct {
	gorm.Model
	Sprint             Sprint
	SprintID           uint `gorm:"not null"`
	Member             userModels.User
	MemberID           uint `gorm:"not null"`
	AllocationPercent  uint `gorm:"not null"`
	ExpectationPercent uint `gorm:"not null"`
	Tasks              []SprintMemberTask
	Vacations          uint                 `gorm:"not null;default:0"`
	Rating             retrospective.Rating `gorm:"default:0; not null"`
	Comment            string               `gorm:"type:text"`
}

// BeforeSave ...
func (sprintMember *SprintMember) BeforeSave(db *gorm.DB) (err error) {
	// Vacations should not be more than sprint length
	if sprintMember.Sprint.EndDate.Sub(*sprintMember.Sprint.StartDate).Hours()/24 > float64(sprintMember.Vacations) {
		err = errors.New("vacations cannot be more than sprint length")
		return err
	}

	return
}
