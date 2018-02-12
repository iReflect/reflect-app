package validators

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v8"
	"reflect"

	retrospectiveSerializers "github.com/iReflect/reflect-app/apps/retrospective/serializers"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
)

// IsValidTaskProviderConfigList validates the TaskProviderConfig
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
			Where("id = ? and active = true", teamID).
			First(&team).Error; err != nil {
			return false
		}
		return true
	}
}
