package services

import (
	"github.com/jinzhu/gorm"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
)

// PermissionService ...
type PermissionService struct {
	DB *gorm.DB
}

// UserCanAccessRetro ...
func (service PermissionService) UserCanAccessRetro(retroID string, userID uint) bool {
	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Joins("JOIN user_teams ON retrospectives.team_id=user_teams.team_id").
		Where("user_teams.user_id=?", userID).
		Where("retrospectives.id=?", retroID).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanAccessSprint ...
func (service PermissionService) UserCanAccessSprint(sprintID string, userID uint) bool {
	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Joins("JOIN user_teams ON retrospectives.team_id=user_teams.team_id").
		Joins("JOIN sprints ON retrospectives.id=sprints.retrospective_id").
		Where("user_teams.user_id=?", userID).
		Where("sprints.id=?", sprintID).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanAccessTask ...
func (service PermissionService) UserCanAccessTask(taskID string, userID uint) bool {
	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Joins("JOIN user_teams ON retrospectives.team_id=user_teams.team_id").
		Joins("JOIN tasks ON retrospectives.id=tasks.retrospective_id").
		Where("user_teams.user_id=?", userID).
		Where("tasks.id=?", taskID).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanEditSprint ...
func (service PermissionService) UserCanEditSprint(sprintID string, userID uint) bool {
	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Joins("JOIN user_teams ON retrospectives.team_id=user_teams.team_id").
		Joins("JOIN sprints ON retrospectives.id=sprints.retrospective_id").
		Where("user_teams.user_id=?", userID).
		Where("sprints.id=?", sprintID).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}
