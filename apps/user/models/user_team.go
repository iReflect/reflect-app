package models

import (
	"time"
	"strconv"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/sirupsen/logrus"
)

var TeamRoleValues = [...]string {
	"Member",
	"Manager",
	"Admin",
}

type TeamRole int8

func (teamRole TeamRole) String() string {
	return TeamRoleValues[teamRole]
}

const (
	MemberRole	TeamRole = iota
	ManagerRole
	AdminRole
)

// UserTeam represent team associations of a users
type UserTeam struct {
	User      User
	UserID    uint `gorm:"primary_key"`
	Team      Team
	TeamID    uint `gorm:"primary_key"`
	Active    bool `gorm:"default:false; not null"`
	Role      TeamRole `gorm:"default:0; not null;type:ENUM(0, 1, 2)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

func RegisterUserTeamToAdmin(Admin *admin.Admin, config admin.Config) {
	userTeam := Admin.AddResource(&UserTeam{}, &config)
	roleMeta := getRoleFieldMeta()
	userTeam.Meta(&roleMeta)
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
