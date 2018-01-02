package validators

import (
	"reflect"
	"gopkg.in/go-playground/validator.v8"

	"github.com/iReflect/reflect-app/apps/feedback/models"
)


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
	return models.ValidateResponseRegex(field.String())
}
