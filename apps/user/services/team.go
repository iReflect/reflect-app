package services

import (
	"errors"

	"github.com/jinzhu/gorm"

	"net/http"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
	"github.com/iReflect/reflect-app/constants"
	"github.com/iReflect/reflect-app/libs/utils"
)

//TeamService ...
type TeamService struct {
	DB *gorm.DB
}

// UserTeamList ...
func (service TeamService) UserTeamList(userID uint, onlyActive bool) (teams *userSerializers.TeamsSerializer, status int, errorCode string, err error) {
	db := service.DB
	teams = new(userSerializers.TeamsSerializer)

	filterQuery := db.Model(&userModels.Team{}).
		Where("teams.deleted_at IS NULL").
		Joins("JOIN user_teams ON teams.id = user_teams.team_id").
		Where("user_teams.user_id = ?", userID).
		Where("teams.active = true").
		Order("teams.name, teams.created_at")

	if onlyActive {
		filterQuery = filterQuery.Where("(leaved_at IS NULL OR leaved_at > NOW())")
	}

	err = filterQuery.Scan(&teams.Teams).Error
	if err != nil {
		responseError := constants.APIErrorMessages[constants.GetUserTeamListError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	return teams, http.StatusOK, "", nil
}

// MemberList ...
func (service TeamService) MemberList(teamID string, userID uint, onlyActive bool, isAdmin bool) (
	members *userSerializers.MembersSerializer, status int, errorCode string, err error) {
	db := service.DB
	members = new(userSerializers.MembersSerializer)
	activeMemberIDs := service.getTeamMemberIDs(teamID, true)
	var memberIDs []uint

	if !isAdmin && !utils.UIntInSlice(userID, activeMemberIDs) {
		responseError := constants.APIErrorMessages[constants.NotTeamMemberError]
		return nil, http.StatusForbidden, responseError.Code, errors.New(responseError.Message)
	}

	if onlyActive {
		memberIDs = activeMemberIDs
	} else {
		memberIDs = service.getTeamMemberIDs(teamID, false)
	}

	err = db.Model(&userModels.User{}).
		Where("users.deleted_at IS NULL").
		Where("id in (?)", memberIDs).
		Order("users.first_name, users.last_name, id").
		Scan(&members.Members).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			responseError := constants.APIErrorMessages[constants.MemberListNotFoundError]
			return nil, http.StatusNotFound, responseError.Code, errors.New(responseError.Message)
		}
		utils.LogToSentry(err)
		responseError := constants.APIErrorMessages[constants.GetRetrospectiveMembersError]
		return nil, http.StatusInternalServerError, responseError.Code, errors.New(responseError.Message)
	}

	return members, http.StatusOK, "", nil
}

func (service TeamService) getTeamMemberIDs(teamID string, onlyActive bool) []uint {
	db := service.DB
	var memberIds []uint

	filterQuery := db.Model(&userModels.UserTeam{}).
		Where("user_teams.deleted_at IS NULL").
		Where("team_id = ?", teamID)
	if onlyActive {
		filterQuery = filterQuery.Where("(leaved_at IS NULL OR leaved_at > NOW())")
	}
	filterQuery.Pluck("user_id", &memberIds)

	return memberIds
}
