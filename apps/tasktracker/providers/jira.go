package providers

import (
	"github.com/iReflect/reflect-app/apps/tasktracker"
	"github.com/iReflect/reflect-app/apps/tasktracker/serializers"
)

//JIRATaskProvider ...
type JIRATaskProvider struct {
}

//JIRATaskConnection ...
type JIRAConnection struct {
	config JIRAConfig
}

type JIRAConfig struct {
	Credentials string
	BaseURL     string
	BoardIds    []string
	JQL         string
}

const (
	TASK_PROVIDER_JIRA_AGILE = "jira"
)

func init() {
	provider := &JIRATaskProvider{}
	tasktracker.RegisterTaskProvider(TASK_PROVIDER_JIRA_AGILE, provider)
}

//Configure ...
func (m *JIRATaskProvider) New(config interface{}) tasktracker.Connection {
	c, ok := config.(JIRAConfig)
	if !ok {
		return nil
	}
	return &JIRAConnection{config: c}
}

//GetTaskList ...
func (m *JIRAConnection) GetTaskList(query string) []serializers.Task {
	return nil
}

//GetSprint ...
func (m *JIRAConnection) GetSprint(sprint string) *serializers.Sprint {
	return nil
}

//GetSprintTaskList ...
func (m *JIRAConnection) GetSprintTaskList(sprint string) []serializers.Task {
	return nil
}
