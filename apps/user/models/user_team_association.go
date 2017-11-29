package models

import "time"

type UserTeamAssociation struct {
	User      User
	UserId    uint       `gorm:"primary_key"`
	Team      Team
	TeamId    uint       `gorm:"primary_key"`
	Active    bool       `gorm:"default:true; not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
