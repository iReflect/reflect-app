package models

import "github.com/jinzhu/gorm"

type FeedbackFormContent struct {
	gorm.Model
	FeedbackForm   FeedbackForm `gorm:"ForeignKey:FeedbackFormID; AssociationForeignKey:ID"`
	FeedbackFormID uint         `gorm:"not null"`
	Skill          Skill        `gorm:"ForeignKey:SkillID; AssociationForeignKey:ID"`
	SkillID        uint         `gorm:"not null"`
	Category       Category     `gorm:"ForeignKey:CategoryID; AssociationForeignKey:ID"`
	CategoryID     uint         `gorm:"not null"`
}
