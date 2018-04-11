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
		Scopes(retroModels.RetroJoinUserTeams).
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
		Scopes(retroModels.RetroJoinSprints, retroModels.RetroJoinUserTeams).
		Where("user_teams.user_id=?", userID).
		Where("retrospectives.id=?", retroID).
		Where("sprints.id=?", sprintID).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanEditSprint ...
func (service PermissionService) UserCanEditSprint(retroID string, sprintID string, userID uint) bool {
	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Scopes(retroModels.RetroJoinSprints, retroModels.RetroJoinUserTeams).
		Where("user_teams.user_id=?", userID).
		Where("retrospectives.id=?", retroID).
		Where("sprints.id=?", sprintID).
		Where("(sprints.status <> ? OR sprints.created_by_id = ?)", retroModels.DraftSprint, userID).
		Where("(sprints.status <> ?)", retroModels.CompletedSprint).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanAccessSprintTask ...
func (service PermissionService) UserCanAccessSprintTask(retroID string, sprintID string, sprintTaskID string, userID uint) bool {
	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Scopes(
			retroModels.RetroJoinSprints,
			retroModels.RetroJoinUserTeams,
			retroModels.SprintJoinST,
			retroModels.STJoinTask).
		Where("user_teams.user_id=?", userID).
		Where("sprint_tasks.id=?", sprintTaskID).
		Where("retrospectives.id=?", retroID).
		Where("sprints.id=?", sprintID).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanEditSprintTask ...
func (service PermissionService) UserCanEditSprintTask(retroID string, sprintID string, sprintTaskID string, userID uint) bool {
	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Scopes(
			retroModels.RetroJoinSprints,
			retroModels.RetroJoinUserTeams,
			retroModels.SprintJoinST,
			retroModels.STJoinTask).
		Where("user_teams.user_id=?", userID).
		Where("sprint_tasks.id=?", sprintTaskID).
		Where("retrospectives.id=?", retroID).
		Where("sprints.id=?", sprintID).
		Where("(sprints.status <> ? OR sprints.created_by_id = ?)", retroModels.DraftSprint, userID).
		Where("(sprints.status <> ?)", retroModels.CompletedSprint).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// CanAccessRetrospectiveFeedback ...
func (service PermissionService) CanAccessRetrospectiveFeedback(sprintID string) bool {
	db := service.DB
	err := db.Model(&retroModels.Sprint{}).
		Where("sprints.id=?", sprintID).
		Where("sprints.status in (?)",
			[]retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroModels.Sprint{}).
		Error
	return err == nil
}
