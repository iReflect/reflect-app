package services

import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"
	"strconv"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	userServices "github.com/iReflect/reflect-app/apps/user/services"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
)

// RetrospectiveService ...
type RetrospectiveService struct {
	DB          *gorm.DB
	TeamService userServices.TeamService
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
func (service RetrospectiveService) Get(retrospectiveID string, isEagerLoading bool) (retrospective *retrospectiveSerializers.Retrospective, err error) {
	db := service.DB

	retrospective = new(retrospectiveSerializers.Retrospective)

	baseQuery := db.Model(&retroModels.Retrospective{})
	if isEagerLoading {
		baseQuery = baseQuery.
			Preload("Team").
			Preload("CreatedBy")
	}

	if err = baseQuery.
		Where("retrospectives.id = ?", retrospectiveID).
		First(&retrospective).Error; err != nil {
		return nil, err
	}
	return retrospective, nil
}

// GetTeamMembers ...
func (service RetrospectiveService) GetTeamMembers(retrospectiveID string, userID uint) (members *userSerializers.MembersSerializer, err error) {
	retro, err := service.Get(retrospectiveID, false)
	if err != nil {
		return nil, err
	}

	members, err = service.TeamService.MemberList(strconv.Itoa(int(retro.TeamID)), userID, true)
	if err != nil {
		return nil, err
	}
	return members, nil
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

// Create the Retrospective with the given values (provided the user is a member of the retrospective's team.
func (service RetrospectiveService) Create(userID uint,
	retrospectiveData *retrospectiveSerializers.RetrospectiveCreateSerializer) (*retroModels.Retrospective, error) {
	db := service.DB
	var err error

	// Check if the user has the permission to create the retro
	if err = db.Model(&userModels.UserTeam{}).
		Where("team_id = ? and user_id = ? and leaved_at IS NULL",
			retrospectiveData.TeamID, userID).
		Find(&userModels.UserTeam{}).Error; err != nil {
		err = errors.New("user doesn't have the permission to create the retro")
		return nil, err
	}

	var retro retroModels.Retrospective
	var taskProviders []byte
	var encryptedTaskProviders []byte

	retro.TeamID = retrospectiveData.TeamID
	retro.CreatedByID = userID
	retro.Title = retrospectiveData.Title
	retro.ProjectName = retrospectiveData.ProjectName
	retro.StoryPointPerWeek = retrospectiveData.StoryPointPerWeek

	if err := tasktracker.ValidateConfigs(retrospectiveData.TaskProviderConfig); err != nil {
		return nil, err
	}

	if taskProviders, err = json.Marshal(retrospectiveData.TaskProviderConfig); err != nil {
		return nil, err
	}

	if encryptedTaskProviders, err = tasktracker.EncryptTaskProviders(taskProviders); err != nil {
		return nil, err
	}
	retro.TaskProviderConfig = encryptedTaskProviders

	err = db.Create(&retro).Error
	return &retro, err
}
