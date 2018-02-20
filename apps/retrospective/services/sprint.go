package services

import (
	"errors"
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/jinzhu/gorm"
)

// SprintService ...
type SprintService struct {
	DB *gorm.DB
}

// DeleteSprint deletes the given sprint
func (service SprintService) DeleteSprint(sprintID string) error {
	db := service.DB
	var sprint retroModels.Sprint
	if err := db.Where("id = ?", sprintID).
		Where("status in (?)", []retroModels.SprintStatus{retroModels.DraftSprint,
			retroModels.ActiveSprint}).
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
	var sprint retroModels.Sprint

	if err := db.Where("id = ?", sprintID).
		Where("status = ?", retroModels.DraftSprint).
		Find(&sprint).Error; err != nil {
		return err
	}

	sprint.Status = retroModels.ActiveSprint
	if rowsAffected := db.Save(&sprint).RowsAffected; rowsAffected == 0 {
		return errors.New("sprint couldn't be activated")
	}
	return nil
}

// Get return details of the given sprint
func (service SprintService) Get(sprintID string, userID uint) (*retrospectiveSerializers.Sprint, error) {
	db := service.DB
	var sprint retrospectiveSerializers.Sprint
	if err := db.Model(&retroModels.Sprint{}).
		Scopes(retroModels.NotDeletedSprint).
		Where("id = ?", sprintID).
		Preload("CreatedBy").
		Find(&sprint).Error; err != nil {
		return nil, err
	}
	return &sprint, nil
}

// GetSprintsList ...
func (service SprintService) GetSprintsList(retrospectiveID string, userID uint) (sprints *retrospectiveSerializers.SprintsSerializer, err error) {
	db := service.DB
	sprints = new(retrospectiveSerializers.SprintsSerializer)

	err = db.Model(&retroModels.Sprint{}).
		Where("retrospective_id = ?", retrospectiveID).
		Where("status in (?) OR (status = (?) AND created_by_id = (?))", []retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}, retroModels.DraftSprint, userID).
		Preload("CreatedBy").
		Order("end_date desc").
		Scan(&sprints.Sprints).Error

	if err != nil {
		return nil, err
	}
	return sprints, nil
}
