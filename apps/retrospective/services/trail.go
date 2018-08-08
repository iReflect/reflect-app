package services

import (
	"github.com/iReflect/reflect-app/constants"
	"github.com/jinzhu/gorm"

	"errors"
	"net/http"
	"strconv"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	trailSerializer "github.com/iReflect/reflect-app/apps/retrospective/serializers"
)

// TrailService ...
type TrailService struct {
	DB *gorm.DB
}

// Add ...
func (service TrailService) Add(action string, actionItem string, actionItemID string, actionByID uint) {
	db := service.DB
	trail := new(retroModels.Trail)

	trail.Action = action
	trail.ActionItem = actionItem
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

	trailSprint := db.Model(&retroModels.Trail{}).
		Where("trails.action_item = ?", constants.ActionItemType[constants.Sprint]).
		Where("trails.action_item_id = ?", sprintID).
		Order("action_item_id").QueryExpr()

	trailSprintMember := db.Model(&retroModels.Trail{}).
		Joins("JOIN sprint_members ON trails.action_item_id = sprint_members.id").
		Where("sprint_members.sprint_id = ?", sprintID).
		Where("trails.action_item = ?", constants.ActionItemType[constants.SprintMember]).
		Order("action_item_id").QueryExpr()

	trailSprintTask := db.Model(&retroModels.Trail{}).
		Joins("JOIN sprint_tasks ON trails.action_item_id = sprint_tasks.id").
		Where("sprint_tasks.sprint_id = ?", sprintID).
		Where("action_item = ?", constants.ActionItemType[constants.SprintTask]).
		Order("action_item_id").QueryExpr()

	trailSprintMemberTask := db.Model(&retroModels.Trail{}).
		Joins("JOIN sprint_member_tasks ON trails.action_item_id = sprint_member_tasks.id").
		Where("trails.action_item = ?", constants.ActionItemType[constants.SprintMemberTask]).
		Joins("JOIN sprint_tasks ON sprint_member_tasks.sprint_task_id = sprint_tasks.id").
		Where("sprint_tasks.sprint_id = ?", sprintID).
		Order("action_item_id").QueryExpr()

	errTrails := db.Raw("SELECT * FROM (?) AS sprint UNION ALL SELECT * FROM (?) AS sprint_member UNION ALL SELECT * FROM (?) AS sprint_task UNION ALL SELECT * FROM (?) AS sprint_member_task",
		trailSprint, trailSprintMember, trailSprintTask, trailSprintMemberTask).
		Scan(&trails.Trails).Error

	if errTrails != nil {
		return nil, http.StatusNoContent, errors.New("No Sprint History")
	}
	return trails, http.StatusOK, nil

}
