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

// Register registers all the validators for the feedback serializers
func (validator RetrospectiveValidators) Register() {
	if err := binding.Validator.RegisterValidation("is_valid_team",
		IsValidTeam(validator.DB)); err != nil {
		logrus.Error(err.Error())
	}

	if err := binding.Validator.RegisterValidation("is_valid_task_provider_config",
		IsValidTaskProviderConfigList); err != nil {
		logrus.Error(err.Error())
	}
}