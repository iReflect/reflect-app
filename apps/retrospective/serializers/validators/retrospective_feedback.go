package validators

import (
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"gopkg.in/go-playground/validator.v8"
	"reflect"
)

// IsValidRetrospectiveFeedbackScope ...
//noinspection GoUnusedParameter
func IsValidRetrospectiveFeedbackScope(
	v *validator.Validate,
	topStruct reflect.Value,
	currentStruct reflect.Value,
	field reflect.Value,
	fieldType reflect.Type,
	fieldKind reflect.Kind,
	param string,
) bool {
	scope := currentStruct.Interface().(*serializers.RetrospectiveFeedbackUpdateSerializer).Scope
	if scope != nil && *scope >= 0 && int(*scope) < len(models.RetrospectiveFeedbackScopeValues) {
		return true
	}
	return false
}
