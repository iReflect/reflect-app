package models

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
)

// Team represent a team/project comprising a set of user
type Team struct {
	gorm.Model
	Name        string `gorm:"type:varchar(64);not null"`
	Description string `gorm:"type:text"`
	Active      bool   `gorm:"default:true; not null"`
	Users       []User
}

// RegisterTeamToAdmin ...
func RegisterTeamToAdmin(Admin *admin.Admin, config admin.Config) {
	team := Admin.AddResource(&Team{}, &config)

	team.IndexAttrs("-Users")
	team.NewAttrs("-Users")
	team.EditAttrs("-Users")
	team.ShowAttrs("-Users")
}
