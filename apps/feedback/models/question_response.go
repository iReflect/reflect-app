package models

import "github.com/jinzhu/gorm"

type QuestionResponse struct {
	gorm.Model
	Feedback              Feedback            `gorm:"ForeignKey:FeedbackID; AssociationForeignKey:ID"`
	FeedbackID            uint                `gorm:"not null"`
	FeedbackFormContent   FeedbackFormContent `gorm:"ForeignKey:FeedbackFormContentID; AssociationForeignKey:ID"`
	FeedbackFormContentID uint                `gorm:"not null"`
	Question              Question            `gorm:"ForeignKey:QuestionID; AssociationForeignKey:ID"`
	QuestionID            uint
	Response              string              `gorm:"type:varchar(100);not null"`
}
