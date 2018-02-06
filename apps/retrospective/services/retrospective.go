package services

import (
	"github.com/jinzhu/gorm"

	retrospectiveModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
)

// RetroSpectiveService ...
type RetroSpectiveService struct {
	DB *gorm.DB
}

// List all the RetroSpectives of all the teams, given user is a member of.
func (service RetroSpectiveService) List(userID uint, perPage int, page int) (
	retroSpectiveList *retrospectiveSerializers.RetroSpectiveListSerializer,
	err error) {
	db := service.DB

	retroSpectiveList = &retrospectiveSerializers.RetroSpectiveListSerializer{}
	retroSpectiveList.RetroSpectives = []retrospectiveSerializers.Retrospective{}

	var offset int
	if perPage < 0 && page > 1 {
		return retroSpectiveList, nil
	} else if page < 1 {
		offset = 0
	} else {
		offset = (page - 1) * perPage
	}

	baseQuery := db.Model(&retrospectiveModels.Retrospective{}).
		Preload("Team").
		Preload("CreatedBy").
		Joins("JOIN user_teams on user_teams.user_id=? AND retrospectives.team_id=user_teams.team_id", userID).
		Limit(perPage)

	if offset != 0 {
		baseQuery = baseQuery.Offset(offset)
	}

	if err := baseQuery.Find(&retroSpectiveList.RetroSpectives).Error; err != nil {
		return nil, err
	}
	return retroSpectiveList, nil
}
