package services

import (
	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/db"
	"strconv"
)

// TrailService ...
type TrailService struct {
}

// Add ...
func (service TrailService) Add(action string, actionItem string, actionItemID string, actionByID uint) {
	db := db.DB
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
