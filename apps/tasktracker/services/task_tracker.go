package services

import (
	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/tasktracker"
	"github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"github.com/iReflect/reflect-app/apps/timetracker"
	userModels "github.com/iReflect/reflect-app/apps/user/models"
	"github.com/iReflect/reflect-app/constants"
)

//TaskTrackerService ...
type TaskTrackerService struct {
	DB *gorm.DB
}

// ConfigList List TaskTracker Configuration
func (service TaskTrackerService) ConfigList() (configList []map[string]interface{}) {
	for _, taskProvider := range tasktracker.TaskProviders {
		configList = append(configList, taskProvider.ConfigTemplate())
	}
	return configList
}

// SupportedTimeTrackersList ...
func (service TaskTrackerService) SupportedTimeTrackersList(taskTracker string, teamID string) (*serializers.TimeProvidersSerializer, error) {
	var timeTrackerList serializers.TimeProvidersSerializer
	var team userModels.Team
	var isGenericTimeTracker bool

	err := service.DB.Model(&userModels.Team{}).Where("id = ?", teamID).Scan(&team).Error
	if err != nil {
		return nil, err
	}
	// check if task tracker also provide time tracking.
	if name, exists := timetracker.TimeProvidersDisplayNameMap[taskTracker]; exists {
		timeTrackerList.TimeProviders = append(timeTrackerList.TimeProviders, serializers.TimeProvider{DisplayName: name, Name: taskTracker})
	}

	for _, genericTimeTracker := range constants.GenericTimeTrackersList {
		if team.TimeProviderName == genericTimeTracker {
			isGenericTimeTracker = true
		} else {
			timeTrackerList.TimeProviders = append(timeTrackerList.TimeProviders, serializers.TimeProvider{
				Name:        genericTimeTracker,
				DisplayName: timetracker.TimeProvidersDisplayNameMap[genericTimeTracker],
			})
		}
	}
	// we will put team task tracker in the first place as a default time provider.
	if taskTracker != team.TimeProviderName && isGenericTimeTracker {
		teamTaskTracker := serializers.TimeProvider{
			Name:        team.TimeProviderName,
			DisplayName: timetracker.TimeProvidersDisplayNameMap[team.TimeProviderName],
		}
		timeTrackerList.TimeProviders = append([]serializers.TimeProvider{teamTaskTracker}, timeTrackerList.TimeProviders...)
	}

	return &timeTrackerList, nil
}
