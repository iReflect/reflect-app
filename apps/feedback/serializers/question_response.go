package serializers

import "github.com/iReflect/reflect-app/db/models/fields"

// QuestionResponseDetail returns the question response for a particular question
type QuestionResponseDetailSerializer struct {
	ID         uint
	Text       string
	Type       int8
	Options    fields.JSONB
	Weight     int
	ResponseID uint
	Response   string
	Comment    string
}
