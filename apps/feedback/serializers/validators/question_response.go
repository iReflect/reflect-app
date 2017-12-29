package validators

import (
	"reflect"
	"regexp"

	validator "gopkg.in/go-playground/validator.v8"
)

// Regex to test the response format of a question
const questionResponseRegexString = "^[0-9]+(,[0-9]+)*$"

var questionResponseRegex = regexp.MustCompile(questionResponseRegexString)

// IsValidQuestionResponse validates the "Response" value of QuestionResponseSerializer
func IsValidQuestionResponse(
	v *validator.Validate,
	topStruct reflect.Value,
	currentStruct reflect.Value,
	field reflect.Value,
	fieldType reflect.Type,
	fieldKind reflect.Kind,
	param string,
) bool {
	questionResponse := field.String()
	// Response can be an empty string only if the feedback is not getting submitted
	return questionResponse == "" || questionResponseRegex.MatchString(questionResponse)
}
