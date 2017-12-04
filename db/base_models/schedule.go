package base_models

import (
	"time"
	"github.com/jinzhu/gorm"
)

type Schedule struct {
	gorm.Model
	Team         Team
	TeamID       uint      `gorm:"not null"`
	PeriodValue  uint      `gorm:"not null"`
	PeriodUnit   string    `gorm:"type:varchar(15); not null"`
	PeriodOffset uint      `gorm:"default:1; not null"`
	ExpireInDays uint      `gorm:"default:10; not null"`
	NextEventAt  time.Time `gorm:"not null"`
	Active       bool      `gorm:"default:true; not null"`
}
