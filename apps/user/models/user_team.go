package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

var TeamRoleValues = [...]string{
	"Member",
	"Manager",
	"Admin",
}

type TeamRole int8

func (teamRole TeamRole) String() string {
	return TeamRoleValues[teamRole]
}

const (
	MemberRole TeamRole = iota
	ManagerRole
	AdminRole
)

// UserTeam represent team associations of a users
type UserTeam struct {
	gorm.Model
	User     User
	UserID   uint `gorm:"not null"`
	Team     Team
	TeamID   uint      `gorm:"not null"`
	Role     TeamRole  `gorm:"default:0; not null"`
	JoinedAt time.Time `gorm:"not null"`
	LeavedAt *time.Time
}

// BeforeSave ...
func (userTeam *UserTeam) BeforeSave(db *gorm.DB) (err error) {

	var userTeams []UserTeam
	var userTeamsCount uint
	userTeamsCount = 0
	baseQuery := db.Where("id <> ? AND user_id = ? AND team_id = ?",
		userTeam.ID,
		userTeam.User.ID,
		userTeam.Team.ID)

	// Provided joined_at-leaved_at duration should not overlap with any other entry for the given user-team pair
	// TODO Add check

	// More than one entries with null leaved_at for given user-team pair should not be allowed
	baseQuery.Where("leaved_at IS NULL").Find(&userTeams).Count(&userTeamsCount)
	if userTeam.LeavedAt == nil && userTeamsCount != 0 {
		err = errors.New("user pre-exist for the team")
		return err
	}

	// Leaved at if provided should be after joined at
	if userTeam.LeavedAt != nil && userTeam.LeavedAt.Before(userTeam.JoinedAt) {
		err = errors.New("leaved at can not be before joined at")
		return err
	}

	return
}

// RegisterUserTeamToAdmin ...
func RegisterUserTeamToAdmin(Admin *admin.Admin, config admin.Config) {
	userTeam := Admin.AddResource(&UserTeam{}, &config)
	roleMeta := getRoleFieldMeta()
	userTeam.Meta(&roleMeta)
	userFieldMeta := GetUserFieldMeta("User")
	userTeam.Meta(&userFieldMeta)
}

// getRoleFieldMeta is the meta config for the user team role field
func getRoleFieldMeta() admin.Meta {
	return admin.Meta{
		Name: "Role",
		Type: "select_one",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			userTeam := value.(*UserTeam)
			return strconv.Itoa(int(userTeam.Role))
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			userTeam := resource.(*UserTeam)
			value, err := strconv.Atoi(metaValue.Value.([]string)[0])
			if err != nil {
				logrus.Error("Cannot convert string to int")
				return
			}
			userTeam.Role = TeamRole(value)
		},
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			for index, value := range TeamRoleValues {
				results = append(results, []string{strconv.Itoa(index), value})
			}
			return
		},
		FormattedValuer: func(value interface{}, context *qor.Context) interface{} {
			userTeam := value.(*UserTeam)
			return userTeam.Role.String()
		},
	}
}
