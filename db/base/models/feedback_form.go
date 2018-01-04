package models

import (
	"github.com/jinzhu/gorm"
	"github.com/iReflect/reflect-app/apps/feedback/models"
)

// FeedbackForm represent template form for feedback
// TODO Add support for versioning
type FeedbackForm struct {
	gorm.Model
	Title       string `gorm:"type:varchar(255); not null"`
	Description string `gorm:"type:text;"`
	Status      models.FeedbackFormStatus   `gorm:"default:0; not null"`
	Archive     bool   `gorm:"default:false; not null"`
}
