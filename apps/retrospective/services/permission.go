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
func (service PermissionService) UserCanAccessSprint(retroID string, sprintID string, userID uint) bool {
	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Joins("JOIN user_teams ON retrospectives.team_id=user_teams.team_id").
		Joins("JOIN sprints ON retrospectives.id=sprints.retrospective_id").
		Where("user_teams.user_id=?", userID).
		Where("retrospectives.id=?", retroID).
		Where("sprints.id=?", sprintID).
		Where("(sprints.status <> ? OR sprints.created_by_id = ?)", retroModels.DraftSprint, userID).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanEditSprint ...
func (service PermissionService) UserCanEditSprint(retroID string, sprintID string, userID uint) bool {
	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Joins("JOIN user_teams ON retrospectives.team_id=user_teams.team_id").
		Joins("JOIN sprints ON retrospectives.id=sprints.retrospective_id").
		Where("user_teams.user_id=?", userID).
		Where("retrospectives.id=?", retroID).
		Where("sprints.id=?", sprintID).
		Where("(sprints.status <> ? OR sprints.created_by_id = ?)", retroModels.DraftSprint, userID).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanAccessTask ...
func (service PermissionService) UserCanAccessTask(retroID string, sprintID string, taskID string, userID uint) bool {
	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Joins("JOIN user_teams ON retrospectives.team_id=user_teams.team_id").
		Joins("JOIN tasks ON retrospectives.id=tasks.retrospective_id").
		Joins("JOIN sprints ON retrospectives.id=sprints.retrospective_id").
		Where("user_teams.user_id=?", userID).
		Where("tasks.id=?", taskID).
		Where("retrospectives.id=?", retroID).
		Where("sprints.id=?", sprintID).
		Where("(sprints.status <> ? OR sprints.created_by_id = ?)", retroModels.DraftSprint, userID).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}
