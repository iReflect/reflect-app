package services

import (
	"errors"
	retrospectiveModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/jinzhu/gorm"
)

// SprintService ...
type SprintService struct {
	DB *gorm.DB
}

// DeleteSprint ...
func (service SprintService) DeleteSprint(sprintID string) error {
	db := service.DB
	var sprint retrospectiveModels.Sprint
	if err := db.Where("id = ?", sprintID).
		Where("status in (?)", []retrospectiveModels.SprintStatus{retrospectiveModels.DraftSprint,
			retrospectiveModels.ActiveSprint}).
		Find(&sprint).Error; err != nil {
		return err
	}
	if rowsAffected := db.Delete(&sprint).RowsAffected; rowsAffected == 0 {
		return errors.New("sprint can't be deleted")
	}
	return nil
}

// ActivateSprint activates the given sprint
func (service SprintService) ActivateSprint(sprintID string) error {
	db := service.DB
	var sprint retrospectiveModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Where("status = ?", retrospectiveModels.DraftSprint).
		Find(&sprint).Error; err != nil {
		return err
	}

	sprint.Status = retrospectiveModels.ActiveSprint
	if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
		return errors.New("sprint couldn't be activated")
	}
	return nil
}
