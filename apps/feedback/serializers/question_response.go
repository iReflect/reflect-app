package serializers

import "github.com/iReflect/reflect-app/apps/feedback/models"

// QuestionResponseDetailSerializer returns the question response for a particular question
type QuestionResponseDetailSerializer struct {
	ID         uint
	Text       string
	Type       models.QuestionType
	Options    interface{}
	Weight     int
	ResponseID uint
	Response   string
	Comment    string
}

// QuestionResponseSerializer returns the question response
type QuestionResponseSerializer struct {
	Response string `json:"response" binding:"is_valid_question_response"`
	Comment  string `json:"comment"`
}
