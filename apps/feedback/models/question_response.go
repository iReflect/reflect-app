package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"regexp"
	"strings"
)

const QuestionResponseSeparator = ","

// Regex to test the response format of a question
const questionResponseRegexString = "^[0-9]+(,[0-9]+)*$"

var questionResponseRegex = regexp.MustCompile(questionResponseRegexString)

// QuestionResponse represent the response/answer to a question asked for a skill
type QuestionResponse struct {
	gorm.Model
	Feedback              Feedback
	FeedbackID            uint `gorm:"not null"`
	FeedbackFormContent   FeedbackFormContent
	FeedbackFormContentID uint `gorm:"not null"`
	Question              Question
	QuestionID            uint   `gorm:"not null"`
	Response              string `gorm:"type:text"`
	Comment               string `gorm:"type:text"`
}

func GetQuestionResponseList(questionResponse string) []string {
	isValid := ValidateResponseRegex(questionResponse)
	if isValid {
		return strings.Split(questionResponse, QuestionResponseSeparator)
	}
	return []string{}
}

func ValidateResponseRegex(questionResponse string) bool {
	// Response can either be an empty string or should match the regex
	return questionResponse == "" || questionResponseRegex.MatchString(questionResponse)
}

func (questionResponse *QuestionResponse) BeforeSave(db *gorm.DB) (err error) {
	// Check if the question response is valid
	question := Question{}
	db.Where("id = ?", questionResponse.QuestionID).First(&question)
	if isValid := question.ValidateQuestionResponse(questionResponse.Response); !isValid {
		err = errors.New("invalid question response")
	}
	return
}
