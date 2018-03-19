package validators

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

//RetrospectiveValidators is used for registering validators for the retrospective app
type RetrospectiveValidators struct {
	DB *gorm.DB
}

// Register registers all the validators for the retrospective serializers
func (validator RetrospectiveValidators) Register() {
	if err := binding.Validator.RegisterValidation("is_valid_team",
		IsValidTeam(validator.DB)); err != nil {
		logrus.Error(err.Error())
	}

	if err := binding.Validator.RegisterValidation("is_valid_task_provider_config",
		IsValidTaskProviderConfigList); err != nil {
		logrus.Error(err.Error())
	}
	if err := binding.Validator.RegisterValidation("is_valid_sprint",
		IsValidSprint); err != nil {
		logrus.Error(err.Error())
	}

	if err := binding.Validator.RegisterValidation("is_valid_rating",
		IsValidRating); err != nil {
		logrus.Error(err.Error())
	}

	if err := binding.Validator.RegisterValidation("is_valid_task_role",
		IsValidTaskRole); err != nil {
		logrus.Error(err.Error())
	}

	if err := binding.Validator.RegisterValidation("is_valid_retrospective_feedback_scope",
		IsValidRetrospectiveFeedbackScope); err != nil {
		logrus.Error(err.Error())
	}
}
