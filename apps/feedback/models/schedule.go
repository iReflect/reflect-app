package models

import (
	"time"
	"github.com/jinzhu/gorm"

	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

type Schedule struct {
	gorm.Model
	Team         userModels.Team `gorm:"ForeignKey:TeamID; AssociationForeignKey:ID"`
	TeamID       uint            `gorm:"not null"`
	PeriodValue  uint            `gorm:"not null"`
	PeriodUnit   string          `gorm:"type:varchar(15); not null"`
	PeriodOffset uint            `gorm:"default:1; not null"`
	Expiry       uint            `gorm:"default:10; not null"`
	NextEventAt  time.Time       `gorm:"not null"`
	Active       bool            `gorm:"default:true; not null"`
}
