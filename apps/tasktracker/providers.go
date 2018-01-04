package tasktracker

import "github.com/iReflect/reflect-app/apps/tasktracker/serializers"

//TaskProvider ...
type TaskProvider interface {
	New(config interface{}) Connection
}

type Connection interface {
	GetTaskList(query string) []serializers.Task
	GetSprint(sprint string) *serializers.Sprint
	GetSprintTaskList(sprint string) []serializers.Task
}

var taskProviders = make(map[string]TaskProvider)

//RegisterTaskProvider ...
func RegisterTaskProvider(name string, newProvider TaskProvider) {
	taskProviders[name] = newProvider
}

//GetTaskProvider ...
func GetTaskProvider(name string) TaskProvider {
	provider, ok := taskProviders[name]
	if ok {
		return provider
	}
	return nil
}
