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
