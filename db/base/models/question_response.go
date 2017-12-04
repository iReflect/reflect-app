package models

import "github.com/jinzhu/gorm"

// QuestionResponse represent the response/answer to a question asked for a skill
type QuestionResponse struct {
	gorm.Model
	Feedback              Feedback
	FeedbackID            uint   `gorm:"not null"`
	FeedbackFormContent   FeedbackFormContent
	FeedbackFormContentID uint   `gorm:"not null"`
	Question              Question
	QuestionID            uint   `gorm:"not null"`
	Response              string `gorm:"type:text"`
	Comment               string `gorm:"type:text"`
}
