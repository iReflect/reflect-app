package models

import (
	"time"
	"github.com/iReflect/reflect-app/apps/user/models"
)

// UserTeam represent team associations of a users
type UserTeam struct {
	User      User
	UserID    uint `gorm:"primary_key"`
	Team      Team
	TeamID    uint `gorm:"primary_key"`
	Role      models.TeamRole `gorm:"default:0; not null;type:ENUM(0, 1, 2)"`
	Active    bool `gorm:"default:false; not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
