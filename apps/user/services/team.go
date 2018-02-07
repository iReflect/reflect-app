package services

import (
	"errors"
	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/libs/utils"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"
)

//TeamService ...
type TeamService struct {
	DB *gorm.DB
}

// MemberList ...
func (service TeamService) MemberList(teamID string, userID uint, onlyActive bool) (members *userSerializers.MembersSerializer, err error) {
	db := service.DB
	members = new(userSerializers.MembersSerializer)
	activeMemberIDs := service.getTeamMemberIDs(teamID, true)
	var memberIDs []uint

	if !utils.UIntInSlice(userID, activeMemberIDs) {
		return nil, errors.New("Must be a member of this team")
	}

	if onlyActive {
		memberIDs = activeMemberIDs
	} else {
	    memberIDs = service.getTeamMemberIDs(teamID, false)
	}

	err = db.Model(&userModels.User{}).Where("id in (?)", memberIDs).Scan(&members.Members).Error
	if err != nil {
		return nil, err
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
