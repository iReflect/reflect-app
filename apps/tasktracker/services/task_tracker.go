package services

import "github.com/iReflect/reflect-app/apps/tasktracker"

//TaskTrackerService ...
type TaskTrackerService struct {
}

//List TaskTracker Configuration
func (service TaskTrackerService) ConfigList() (configList []map[string]interface{}) {
	for _, taskProvider := range tasktracker.TaskProviders {
		configList = append(configList, taskProvider.ConfigTemplate())
	}
	return configList
}
