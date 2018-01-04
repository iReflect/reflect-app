package models

import (
	"time"
)

// UserTeam represent team associations of a users
type UserTeam struct {
	User      User
	UserID    uint `gorm:"primary_key"`
	Team      Team
	TeamID    uint `gorm:"primary_key"`
	Role      int8 `gorm:"default:0; not null"`
	Active    bool `gorm:"default:false; not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
