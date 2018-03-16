package services

import (
	"errors"
	"github.com/jinzhu/gorm"

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
func (service TeamService) UserTeamList(userID uint, onlyActive bool) (*userSerializers.TeamsSerializer, error) {
	db := service.DB
	teams := new(userSerializers.TeamsSerializer)

	filterQuery := db.Model(&userModels.Team{}).
		Joins("JOIN user_teams ON teams.id = user_teams.team_id").
		Where("user_teams.user_id = ?", userID).
		Where("teams.active = true")

	if onlyActive {
		filterQuery = filterQuery.Where("(leaved_at IS NULL OR leaved_at > NOW())")
	}

	err := filterQuery.Scan(&teams.Teams).Error
	if err != nil {
		return nil, errors.New(constants.NoUserTeamFound)
	}
	return teams, nil
}

// MemberList ...
func (service TeamService) MemberList(teamID string, userID uint, onlyActive bool) (*userSerializers.MembersSerializer, error) {
	db := service.DB
	members := new(userSerializers.MembersSerializer)
	activeMemberIDs := service.getTeamMemberIDs(teamID, true)
	var memberIDs []uint

	if !utils.UIntInSlice(userID, activeMemberIDs) {
		return nil, errors.New(constants.UserNotATeamMember)
	}

	if onlyActive {
		memberIDs = activeMemberIDs
	} else {
		memberIDs = service.getTeamMemberIDs(teamID, false)
	}

	err := db.Model(&userModels.User{}).Where("id in (?)", memberIDs).Scan(&members.Members).Error
	if err != nil {
		return nil, errors.New(constants.NoTeamMemberFound)
	}

	return members, nil
}

func (service TeamService) getTeamMemberIDs(teamID string, onlyActive bool) []uint {
	db := service.DB
	var memberIds []uint

	filterQuery := db.Model(&userModels.UserTeam{}).Where("team_id = ?", teamID)
	if onlyActive {
		filterQuery = filterQuery.Where("(leaved_at IS NULL OR leaved_at > NOW())")
	}
	filterQuery.Pluck("user_id", &memberIds)

	return memberIds
}
