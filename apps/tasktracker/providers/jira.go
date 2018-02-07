package providers

import (
	"github.com/iReflect/reflect-app/apps/tasktracker"
	"github.com/iReflect/reflect-app/apps/tasktracker/serializers"

	"github.com/andygrunwald/go-jira"
)

//JIRATaskProvider ...
type JIRATaskProvider struct {
}

//JIRAConnection ...
type JIRAConnection struct {
	config JIRAConfig
}

//JIRAConfig ...
type JIRAConfig struct {
	Credentials string
	BaseURL     string
	BoardIds    []string
	JQL         string
}

// TaskProviderJiraAgile ...
const (
	TaskProviderJiraAgile = "jira"
)

func init() {
	provider := &JIRATaskProvider{}
	tasktracker.RegisterTaskProvider(TaskProviderJiraAgile, provider)
}

//New ...
func (m *JIRATaskProvider) New(config interface{}) tasktracker.Connection {
	c, ok := config.(JIRAConfig)
	if !ok {
		return nil
	}
	return &JIRAConnection{config: c}
}

//GetTaskList ...
func (m *JIRAConnection) GetTaskList(query string) []serializers.Task {
	jira.NewClient(nil, "")
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
