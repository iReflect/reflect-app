package services

import (
	"github.com/jinzhu/gorm"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
)

// RetrospectiveService ...
type RetrospectiveService struct {
	DB *gorm.DB
}

// List all the Retrospectives of all the teams, given user is a member of.
func (service RetrospectiveService) List(userID uint, perPage int, page int) (
	retrospectiveList *retrospectiveSerializers.RetrospectiveListSerializer,
	err error) {
	db := service.DB

	retrospectiveList = &retrospectiveSerializers.RetrospectiveListSerializer{}
	retrospectiveList.Retrospectives = []retrospectiveSerializers.Retrospective{}

	var offset int
	if perPage < 0 && page > 1 {
		return retrospectiveList, nil
	} else if page < 1 {
		offset = 0
	} else {
		offset = (page - 1) * perPage
	}

	baseQuery := db.Model(&retroModels.Retrospective{}).
		Preload("Team").
		Preload("CreatedBy").
		Joins("JOIN user_teams ON user_teams.user_id=? AND retrospectives.team_id=user_teams.team_id", userID).
		Limit(perPage)

	if offset != 0 {
		baseQuery = baseQuery.Offset(offset)
	}

	if err := baseQuery.Find(&retrospectiveList.Retrospectives).Error; err != nil {
		return nil, err
	}
	return retrospectiveList, nil
}

// UserCanAccessRetro ...
func (service RetrospectiveService) UserCanAccessRetro(retroID string, userID uint) bool {
	db := service.DB
	exists := db.Model(&retroModels.Retrospective{}).
		Joins("JOIN user_teams ON retrospectives.team_id=user_teams.team_id").
		Where("user_teams.user_id=?", userID).
		Where("retrospectives.id=?", retroID).
		First(&retrospectiveSerializers.Retrospective{}).
		RecordNotFound()
	return !exists
}
