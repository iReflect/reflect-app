package services

import (
	"github.com/jinzhu/gorm"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	"strconv"
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
