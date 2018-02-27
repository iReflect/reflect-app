package validators

import (
	retroSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"gopkg.in/go-playground/validator.v8"
	"reflect"
)

func IsValidSprint(
	v *validator.Validate,
	topStruct reflect.Value,
	currentStruct reflect.Value,
	field reflect.Value,
	fieldType reflect.Type,
	fieldKind reflect.Kind,
	param string,
) bool {
	sprintID := currentStruct.Interface().(*retroSerializers.CreateSprintSerializer).SprintID
	startDate := currentStruct.Interface().(*retroSerializers.CreateSprintSerializer).StartDate
	endDate := currentStruct.Interface().(*retroSerializers.CreateSprintSerializer).EndDate

	if sprintID == "" && (startDate == nil || endDate == nil) {
		return false
	}
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return false
	}
	return true
}
