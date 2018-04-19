package validators

import (
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v8"
	"reflect"

	"github.com/iReflect/reflect-app/apps/retrospective"
	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

// IsValidTaskProviderConfigList validates the TaskProviderConfig
//noinspection GoUnusedParameter
func IsValidTaskProviderConfigList(
	v *validator.Validate,
	topStruct reflect.Value,
	currentStruct reflect.Value,
	field reflect.Value,
	fieldType reflect.Type,
	fieldKind reflect.Kind,
	param string,
) bool {
	var isValid bool
	var providerType string
	var providerData, credentials map[string]interface{}
	taskProviderConfigList := currentStruct.Interface().(*retrospectiveSerializers.RetrospectiveCreateSerializer).TaskProviderConfig
	// There should be at least one task provider config
	if len(taskProviderConfigList) == 0 {
		return false
	}
	for _, taskProviderConfig := range taskProviderConfigList {
		if providerType, isValid = taskProviderConfig["type"].(string); !isValid {
			return false
		}

		if providerData, isValid = taskProviderConfig["data"].(map[string]interface{}); !isValid {
			return false
		}

		for _, taskType := range tasktracker.TaskTypes {
			_, isValid := providerData[taskType].(string)
			if !isValid {
				return false
			}
		}

		if credentials, isValid = providerData["credentials"].(map[string]interface{}); !isValid {
			return false
		}

		if err := tasktracker.ValidateCredentials(credentials); err != nil {
			return false
		}

		taskProvider := tasktracker.GetTaskProvider(providerType)
		if taskProvider == nil {
			return false
		}
		taskProviderConnection := taskProvider.New(providerData)
		if taskProviderConnection == nil {
			return false
		}
	}
	return true
}

// IsValidTeam validates the Team, given the team id, it checks if the team exists and the user is a team member
func IsValidTeam(db *gorm.DB) validator.Func {
	return func(
		v *validator.Validate,
		topStruct reflect.Value,
		currentStruct reflect.Value,
		field reflect.Value,
		fieldType reflect.Type,
		fieldKind reflect.Kind,
		param string,
	) bool {
		var team userModels.Team
		teamID := currentStruct.Interface().(*retrospectiveSerializers.RetrospectiveCreateSerializer).TeamID
		if err := db.Model(&userModels.Team{}).
			Where("deleted_at IS NULL").
			Where("id = ? and active = true", teamID).
			First(&team).Error; err != nil {
			return false
		}
		return true
	}
}

// IsValidRating ...
//noinspection GoUnusedParameter
func IsValidRating(
	v *validator.Validate,
	topStruct reflect.Value,
	currentStruct reflect.Value,
	field reflect.Value,
	fieldType reflect.Type,
	fieldKind reflect.Kind,
	param string,
) bool {
	rating := currentStruct.Interface().(*retrospectiveSerializers.SprintTaskMemberUpdate).Rating
	if rating != nil && *rating >= 0 && int(*rating) < len(retrospective.RatingValues) {
		return true
	}
	return false
}

// IsValidTaskRole ...
//noinspection GoUnusedParameter
func IsValidTaskRole(
	v *validator.Validate,
	topStruct reflect.Value,
	currentStruct reflect.Value,
	field reflect.Value,
	fieldType reflect.Type,
	fieldKind reflect.Kind,
	param string,
) bool {
	role := currentStruct.Interface().(*retrospectiveSerializers.SprintTaskMemberUpdate).Role
	if role != nil && *role >= 0 && int(*role) < len(models.MemberTaskRoleValues) {
		return true
	}
	return false
}
