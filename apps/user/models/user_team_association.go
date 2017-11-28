package models

import "github.com/jinzhu/gorm"

type UserTeamAssociation struct {
	gorm.Model
	User   User `gorm:"ForeignKey:UserID;AssociationForeignKey:ID"`
	UserId uint `gorm:"not null"`
	Team   Team `gorm:"ForeignKey:TeamID;AssociationForeignKey:ID"`
	TeamId uint `gorm:"not null"`
	Active bool `gorm:"default:true; not null"`
}
