package services

import (
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/db"
)

// PermissionService ...
type PermissionService struct {
}

// UserCanAccessRetro ...
func (service PermissionService) UserCanAccessRetro(retroID string, userID uint) bool {
	db := db.DB
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
	db := db.DB
	err := db.Model(&retroModels.Retrospective{}).
		Scopes(retroModels.RetroJoinSprints, retroModels.RetroJoinUserTeams).
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
	db := db.DB
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

// UserCanAccessTask ...
func (service PermissionService) UserCanAccessTask(retroID string, sprintID string, taskID string, userID uint) bool {
	db := db.DB
	err := db.Model(&retroModels.Retrospective{}).
		Scopes(retroModels.RetroJoinSprints, retroModels.RetroJoinTasks, retroModels.RetroJoinUserTeams).
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

// UserCanEditTask ...
func (service PermissionService) UserCanEditTask(retroID string, sprintID string, taskID string, userID uint) bool {
	db := db.DB
	err := db.Model(&retroModels.Retrospective{}).
		Scopes(retroModels.RetroJoinSprints, retroModels.RetroJoinTasks, retroModels.RetroJoinUserTeams).
		Where("user_teams.user_id=?", userID).
		Where("tasks.id=?", taskID).
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
	db := db.DB
	err := db.Model(&retroModels.Sprint{}).
		Where("sprints.id=?", sprintID).
		Where("sprints.status in (?)",
			[]retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}).
		Scopes(retroModels.NotDeletedSprint).
		Find(&retroModels.Sprint{}).
		Error
	return err == nil
}
