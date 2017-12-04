package models

import "github.com/jinzhu/gorm"

// FeedbackFormContent represent the content of the feedback form
// TODO Add support for making it non-editable
type FeedbackFormContent struct {
	gorm.Model
	FeedbackForm   FeedbackForm
	FeedbackFormID uint `gorm:"not null"`
	Skill          Skill
	SkillID        uint `gorm:"not null"`
	Category       Category
	CategoryID     uint `gorm:"not null"`
}
