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
		Joins("JOIN user_teams on user_teams.user_id = ? AND"+
			" retrospectives.team_id = user_teams.team_id", userID).
		Preload("Team").
		Preload("CreatedBy").
		Limit(perPage).
		Order("created_at desc")

	if offset != 0 {
		baseQuery = baseQuery.Offset(offset)
	}

	if err := baseQuery.Find(&retrospectiveList.Retrospectives).Error; err != nil {
		return nil, err
	}
	return retrospectiveList, nil
}

// Get the details of the given RetroSpective.
func (service RetrospectiveService) Get(retrospectiveID string) (retrospective *retrospectiveSerializers.Retrospective, err error) {
	db := service.DB

	retrospective = new(retrospectiveSerializers.Retrospective)

	if err = db.Model(&retroModels.Retrospective{}).
		Preload("Team").
		Preload("CreatedBy").
		Where("retrospectives.id = ?", retrospectiveID).
		First(&retrospective).Error; err != nil {
		return nil, err
	}
	return retrospective, nil
}

// GetLatestSprint returns the latest sprint for the retro
func (service RetrospectiveService) GetLatestSprint(retroID string) (*retrospectiveSerializers.Sprint, error) {
	db := service.DB
	var sprint retrospectiveSerializers.Sprint
	if err := db.Model(&retroModels.Sprint{}).
		Where("retrospective_id = ?", retroID).
		Where("status in (?)", []retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}).
		Order("end_date desc").
		Preload("CreatedBy").
		First(&sprint).Error; err != nil {
		return nil, err
	}
	return &sprint, nil
}
