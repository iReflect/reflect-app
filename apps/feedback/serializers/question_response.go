package serializers

// QuestionResponseDetailSerializer returns the question response for a particular question
type QuestionResponseDetailSerializer struct {
	ID         uint
	Text       string
	Type       int8
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
