package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// UserTeam represent team associations of a users
type UserTeam struct {
	gorm.Model
	User     User
	UserID   uint `gorm:"not null"`
	Team     Team
	TeamID   uint      `gorm:"not null"`
	Role     int8      `gorm:"default:0; not null"`
	JoinedAt time.Time `gorm:"not null"`
	LeavedAt *time.Time
}
