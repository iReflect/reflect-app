package services

import (
	"github.com/jinzhu/gorm"

	"github.com/iReflect/reflect-app/apps/tasktracker"
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
func (service TaskTrackerService) SupportedTimeTrackersList(taskTracker string, teamID string) ([]string, error) {
	var timeTrackerList []string
	var team userModels.Team
	var isGenricTimeTracker bool

	err := service.DB.Model(&userModels.Team{}).Where("id = ?", teamID).Scan(&team).Error
	if err != nil {
		return []string{}, err
	}
	// check if task tracker also provide time
	if name, exists := timetracker.TimeProvidersDisplayNameMap[taskTracker]; exists {
		timeTrackerList = append(timeTrackerList, name)
	}

	for _, genricTimetracker := range constants.GenricTimeTrackersList {
		if team.TimeProviderName == genricTimetracker {
			isGenricTimeTracker = true
			continue
		}
		timeTrackerList = append(timeTrackerList, timetracker.TimeProvidersDisplayNameMap[genricTimetracker])
	}
	if taskTracker != team.TimeProviderName && isGenricTimeTracker {
		teamTaskTracker := timetracker.TimeProvidersDisplayNameMap[team.TimeProviderName]
		timeTrackerList = append([]string{teamTaskTracker}, timeTrackerList...)
	}

	return timeTrackerList, nil
}
