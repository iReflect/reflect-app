package providers

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/andygrunwald/go-jira"

	"github.com/iReflect/reflect-app/apps/tasktracker"
	"github.com/iReflect/reflect-app/apps/tasktracker/serializers"
)

// JIRATaskProvider ...
type JIRATaskProvider struct {
}

// JIRAConnection ...
type JIRAConnection struct {
	config JIRAConfig
	client *jira.Client
}

// JIRAConfig ...
type JIRAConfig struct {
	Credentials tasktracker.Credentials
	BaseURL     string `json:"BaseURL"`
	BoardIds    string `json:"BoardIds"`
	JQL         string `json:"JQL"`
}

// TaskProviderJira ...
const (
	TaskProviderJira = "jira"
)

func init() {
	provider := &JIRATaskProvider{}
	tasktracker.RegisterTaskProvider(TaskProviderJira, provider)
}

// New ...
func (p *JIRATaskProvider) New(config interface{}) tasktracker.Connection {
	var c JIRAConfig

	switch config.(type) {
	case []byte:
		c = JIRAConfig{}
		err := json.Unmarshal(config.([]byte), &c)

		if err != nil {
			return nil
		}
	case JIRAConfig:
		c = config.(JIRAConfig)
	default:
		return nil
	}

	client, err := jira.NewClient(nil, c.BaseURL)

	if err != nil {
		return nil
	}

	switch c.Credentials.Type {
	case "basicAuth":
		client.Authentication.SetBasicAuth(c.Credentials.Username, c.Credentials.Password)
	default:
		return nil
	}

	return &JIRAConnection{config: c, client: client}
}

// GetTaskList ...
func (c *JIRAConnection) GetTaskList(query string) []serializers.Task {
	return nil
}

// GetSprint ...
func (c *JIRAConnection) GetSprint(sprint string) *serializers.Sprint {
	return nil
}

// GetSprintTaskList ...
func (c *JIRAConnection) GetSprintTaskList(sprint string) []serializers.Task {
	return nil
}

// ValidateConfig ...
func (c *JIRAConnection) ValidateConfig() error {
	boardIDs := strings.Split(c.config.BoardIds, ",")

	for _, boardID := range boardIDs {
		boardID, err := strconv.Atoi(boardID)
		if err != nil {
			return err
		}
		_, _, err = c.client.Board.GetBoard(boardID)
		if err != nil {
			return err
		}
	}
	return nil
}
