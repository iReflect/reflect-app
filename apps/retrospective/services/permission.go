package services

import (
	"github.com/jinzhu/gorm"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

// PermissionService ...
type PermissionService struct {
	DB *gorm.DB
}

// UserCanAccessRetro ...
func (service PermissionService) UserCanAccessRetro(retroID string, userID uint) bool {
	if service.IsUserAdmin(userID) {
		return true
	}

	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Where("retrospectives.deleted_at IS NULL").
		Scopes(retroModels.RetroJoinUserTeams).
		Where("user_teams.user_id = ?", userID).
		Where("retrospectives.id = ?", retroID).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanCreateOrEditRetro ...
func (service PermissionService) UserCanCreateOrEditRetro(teamID uint, userID uint) bool {
	if service.IsUserAdmin(userID) {
		return true
	}
	db := service.DB
	// Check if the user has the permission to create or update the retro
	err := db.Model(&userModels.UserTeam{}).
		Where("user_teams.deleted_at IS NULL").
		Where("team_id = ? and user_id = ? and leaved_at IS NULL", teamID, userID).
		Find(&userModels.UserTeam{}).Error
	return err == nil
}

// UserCanAccessSprint ...
func (service PermissionService) UserCanAccessSprint(retroID string, sprintID string, userID uint) bool {
	if service.IsUserAdmin(userID) {
		return true
	}

	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Where("retrospectives.deleted_at IS NULL").
		Scopes(retroModels.RetroJoinSprints, retroModels.RetroJoinUserTeams).
		Where("user_teams.user_id = ?", userID).
		Where("retrospectives.id = ?", retroID).
		Where("sprints.id = ?", sprintID).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanEditSprint ...
func (service PermissionService) UserCanEditSprint(retroID string, sprintID string, userID uint) bool {
	if service.IsUserAdmin(userID) {
		return true
	}

	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Where("retrospectives.deleted_at IS NULL").
		Scopes(retroModels.RetroJoinSprints, retroModels.RetroJoinUserTeams).
		Where("user_teams.user_id = ?", userID).
		Where("retrospectives.id = ?", retroID).
		Where("sprints.id = ?", sprintID).
		Where("(sprints.status <> ? OR sprints.created_by_id = ?)", retroModels.DraftSprint, userID).
		Where("(sprints.status <> ?)", retroModels.CompletedSprint).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanAccessSprintTask ...
func (service PermissionService) UserCanAccessSprintTask(retroID string, sprintID string, sprintTaskID string, userID uint) bool {
	if service.IsUserAdmin(userID) {
		return true
	}

	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Where("retrospectives.deleted_at IS NULL").
		Scopes(
			retroModels.RetroJoinSprints,
			retroModels.RetroJoinUserTeams,
			retroModels.SprintJoinST,
			retroModels.STJoinTask).
		Where("user_teams.user_id = ?", userID).
		Where("sprint_tasks.id = ?", sprintTaskID).
		Where("retrospectives.id = ?", retroID).
		Where("sprints.id = ?", sprintID).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// UserCanEditSprintTask ...
func (service PermissionService) UserCanEditSprintTask(retroID string, sprintID string, sprintTaskID string, userID uint) bool {
	if service.IsUserAdmin(userID) {
		return true
	}

	db := service.DB
	err := db.Model(&retroModels.Retrospective{}).
		Where("retrospectives.deleted_at IS NULL").
		Scopes(
			retroModels.RetroJoinSprints,
			retroModels.RetroJoinUserTeams,
			retroModels.SprintJoinST,
			retroModels.STJoinTask).
		Where("user_teams.user_id = ?", userID).
		Where("sprint_tasks.id = ?", sprintTaskID).
		Where("retrospectives.id = ?", retroID).
		Where("sprints.id = ?", sprintID).
		Where("(sprints.status <> ? OR sprints.created_by_id = ?)", retroModels.DraftSprint, userID).
		Where("(sprints.status <> ?)", retroModels.CompletedSprint).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroSerializers.Retrospective{}).
		Error
	return err == nil
}

// CanAccessRetrospectiveFeedback ...
func (service PermissionService) CanAccessRetrospectiveFeedback(sprintID string, userID uint) bool {
	if service.IsUserAdmin(userID) {
		return true
	}

	db := service.DB
	err := db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Where("sprints.id = ?", sprintID).
		Where("sprints.status in (?)",
			[]retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroModels.Sprint{}).
		Error
	return err == nil
}

// IsUserAdmin ...
func (service PermissionService) IsUserAdmin(userID uint) bool {
	db := service.DB
	err := db.Model(&userModels.User{}).
		Where("users.id = ?", userID).
		Where("users.is_admin = ?", true).
		Find(&userModels.User{}).
		Error
	return err == nil
}
