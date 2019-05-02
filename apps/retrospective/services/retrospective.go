package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
	userServices "github.com/iReflect/reflect-app/apps/user/services"
	"github.com/iReflect/reflect-app/constants"
	"github.com/iReflect/reflect-app/libs/utils"
)

// RetrospectiveService ...
type RetrospectiveService struct {
	DB          *gorm.DB
	TeamService userServices.TeamService
}

// List all the Retrospectives of all the teams, given user is a member of.
func (service RetrospectiveService) List(userID uint, perPageString string, pageString string, isAdmin bool) (
	retrospectiveList *retroSerializers.RetrospectiveListSerializer,
	status int,
	errorCode string,
	err error) {
	db := service.DB

	retrospectiveList = &retroSerializers.RetrospectiveListSerializer{}
	retrospectiveList.Retrospectives = []retroSerializers.Retrospective{}

	perPage, err := strconv.Atoi(perPageString)
	if err != nil {
		perPage = -1
	}
	page, err := strconv.Atoi(pageString)
	if err != nil {
		page = 1
	}

	var offset int
	if perPage < 0 && page > 1 {
		return retrospectiveList, http.StatusNoContent, "", nil
	} else if page < 1 {
		offset = 0
	} else {
		offset = (page - 1) * perPage
	}

	baseQuery := db.Model(&retroModels.Retrospective{}).
		Where("retrospectives.deleted_at IS NULL").
		Preload("CreatedBy").
		Order("created_at DESC, title, id").
		Limit(perPage).
		Offset(offset).
		Select("DISTINCT(retrospectives.*)")

	if isAdmin {
		err = baseQuery.
			Find(&retrospectiveList.Retrospectives).
			Error

	} else {
		err = baseQuery.
			Scopes(retroModels.RetroJoinUserTeams).
			Where("user_teams.user_id = ?", userID).
			Preload("Team").
			Find(&retrospectiveList.Retrospectives).
			Error
	}

	if err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.RetrospectiveListError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)

	}
	return retrospectiveList, http.StatusOK, "", nil
}

// Get the details of the given Retrospective.
func (service RetrospectiveService) Get(retroID string, isEagerLoading bool) (
	retro *retroSerializers.Retrospective,
	status int,
	errorCode string,
	err error) {
	db := service.DB

	retro = new(retroSerializers.Retrospective)

	baseQuery := db.Model(&retroModels.Retrospective{}).
		Where("retrospectives.deleted_at IS NULL")
	if isEagerLoading {
		baseQuery = baseQuery.
			Preload("Team").
			Preload("CreatedBy")
	}

	err = baseQuery.
		Where("retrospectives.id = ?", retroID).
		First(&retro).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.RetrospectiveNotFoundError]
			return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.RetrospectiveDetailsError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	return retro, http.StatusOK, "", nil
}

// GetTeamMembers ...
func (service RetrospectiveService) GetTeamMembers(retrospectiveID string, userID uint, isAdmin bool) (
	members *userSerializers.MembersSerializer, status int, errorCode string, err error) {
	retro, status, errorCode, err := service.Get(retrospectiveID, false)
	if err != nil {
		return nil, status, errorCode, err
	}

	members, status, errorCode, err = service.TeamService.MemberList(strconv.Itoa(int(retro.TeamID)), userID, true, isAdmin)
	if err != nil {
		return nil, status, errorCode, err
	}

	return members, http.StatusOK, "", nil
}

// GetLatestSprint returns the latest sprint for the retro
func (service RetrospectiveService) GetLatestSprint(retroID string, userID uint) (*retroSerializers.Sprint, int, string, error) {
	db := service.DB
	var sprint retroSerializers.Sprint

	err := db.Model(&retroModels.Sprint{}).
		Where("sprints.deleted_at IS NULL").
		Where("retrospective_id = ?", retroID).
		Where("status in (?)", []retroModels.SprintStatus{retroModels.ActiveSprint, retroModels.CompletedSprint}).
		Order("end_date DESC").
		Preload("CreatedBy").
		First(&sprint).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.RetrospectiveNoSprintError]
			return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.RetrospectiveLatestSprintError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	sprint.SetEditable(userID)
	return &sprint, http.StatusOK, "", nil
}

// Create the Retrospective with the given values (provided the user is a member of the retrospective's team.
func (service RetrospectiveService) Create(userID uint,
	retrospectiveData *retroSerializers.RetrospectiveCreateSerializer) (*retroModels.Retrospective, int, string, error) {
	db := service.DB
	var err error

	// Check if the user has the permission to create the retro
	err = db.Model(&userModels.UserTeam{}).
		Where("user_teams.deleted_at IS NULL").
		Where("team_id = ? and user_id = ? and leaved_at IS NULL",
			retrospectiveData.TeamID, userID).
		Find(&userModels.UserTeam{}).Error
	if err != nil {
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.CreateRetrospectivePermissionError]
		return nil, http.StatusForbidden, responseError.Code, errors.New(responseError.Message)
	}

	var retro retroModels.Retrospective
	var taskProviders []byte
	var encryptedTaskProviders []byte

	retro.TeamID = retrospectiveData.TeamID
	retro.CreatedByID = userID
	retro.Title = retrospectiveData.Title
	retro.ProjectName = retrospectiveData.ProjectName
	retro.TimeProviderName = retrospectiveData.TimeProviderName
	retro.StoryPointPerWeek = retrospectiveData.StoryPointPerWeek

	if err := tasktracker.ValidateConfigs(retrospectiveData.TaskProviderConfig); err != nil {
		return nil, http.StatusBadRequest, "", err
	}

	responseError := constants.APIErrorMessages[constants.CreateRetrospectiveError]
	if taskProviders, err = json.Marshal(retrospectiveData.TaskProviderConfig); err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	if encryptedTaskProviders, err = tasktracker.EncryptTaskProviders(taskProviders); err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	retro.TaskProviderConfig = encryptedTaskProviders

	err = db.Create(&retro).Error
	if err != nil {
		utils.LogToSentry(err)
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}
	return &retro, http.StatusCreated, "", nil
}
