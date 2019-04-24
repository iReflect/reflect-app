package services

import (
	"github.com/jinzhu/gorm"

	"errors"
	"net/http"
	"strconv"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	trailSerializer "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/constants"
)

// TrailService ...
type TrailService struct {
	DB *gorm.DB
}

// Add ...
func (service TrailService) Add(action constants.ActionType, actionItem constants.ActionItemType, actionItemID string, actionByID uint) {
	db := service.DB
	trail := new(retroModels.Trail)

	trail.Action = constants.ActionTypeMap[action]
	trail.ActionItem = constants.ActionItemTypeMap[actionItem]
	intID, err := strconv.Atoi(actionItemID)
	if err != nil {
		return
	}
	trail.ActionItemID = uint(intID)
	trail.ActionByID = actionByID

	db.Create(&trail)
	return
}

// GetTrails method to get history of trails for a particular sprint
func (service TrailService) GetTrails(sprintID uint) (trails *trailSerializer.TrailSerializer, status int, err error) {
	db := service.DB
	trails = new(trailSerializer.TrailSerializer)

	sprintTrail := db.Model(&retroModels.Trail{}).
		Where("trails.action_item = ?", constants.ActionItemTypeMap[constants.Sprint]).
		Where("trails.action_item_id = ?", sprintID).
		QueryExpr()

	sprintMemberTrail := db.Model(&retroModels.Trail{}).
		Scopes(retroModels.TrailJoinSM).
		Where("sprint_members.sprint_id = ?", sprintID).
		Where("trails.action_item = ?", constants.ActionItemTypeMap[constants.SprintMember]).
		QueryExpr()

	sprintTaskTrail := db.Model(&retroModels.Trail{}).
		Scopes(retroModels.TrailJoinST).
		Where("sprint_tasks.sprint_id = ?", sprintID).
		Where("trails.action_item = ?", constants.ActionItemTypeMap[constants.SprintTask]).
		QueryExpr()

	sprintMemberTaskTrail := db.Model(&retroModels.Trail{}).
		Scopes(retroModels.TrailJoinSMT, retroModels.SMTJoinST).
		Where("trails.action_item = ?", constants.ActionItemTypeMap[constants.SprintMemberTask]).
		Where("sprint_tasks.sprint_id = ?", sprintID).
		QueryExpr()

	retroFeedbackTrail := db.Model(&retroModels.Trail{}).
		Scopes(retroModels.TrailJoinFeedback).
		Where("trails.action_item = ?", constants.ActionItemTypeMap[constants.RetrospectiveFeedback]).
		QueryExpr()

	err = db.Raw(
		`SELECT * FROM (?) AS sprint_trails
		UNION SELECT * FROM (?) AS sprint_member_trails
		UNION SELECT * FROM (?) AS sprint_task_trails
		UNION SELECT * FROM (?) AS sprint_member_task_trails
		UNION SELECT * FROM (?) AS retro_feedback_trails`,
		sprintTrail, sprintMemberTrail, sprintTaskTrail, sprintMemberTaskTrail, retroFeedbackTrail).
		Preload("ActionBy").
		Order("created_at DESC").
		Find(&trails.Trails).Error

	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("failed to get the sprint trails")
	}
	return trails, http.StatusOK, nil

}
