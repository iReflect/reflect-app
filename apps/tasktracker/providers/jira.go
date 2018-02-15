package providers

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/andygrunwald/go-jira"

	"github.com/iReflect/reflect-app/apps/tasktracker"
	"github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
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
	Credentials   tasktracker.Credentials
	BaseURL       string `json:"BaseURL"`
	BoardIds      string `json:"BoardIds"`
	JQL           string `json:"JQL"`
	EstimateField string `json:"EstimateField"`
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

// ConfigTemplate ...
func (p *JIRATaskProvider) ConfigTemplate() map[string]interface{} {
	template := `{
      "Type": "jira",
      "DisplayTitle": "JIRA",
      "SupportedAuthTypes": ["basicAuth"],
      "Fields": [
        {
          "FieldName": "BaseURL",
          "FieldDisplayName": "Base URL of the project. eg. 'ireflect.atlassian.net'",
          "Type": "string",
          "Required": true
        },
        {
          "FieldName": "BoardIds",
          "FieldDisplayName": "Board IDs (Comma Separated)",
          "Type": "string",
          "Required": true
        },
        {
          "FieldName": "JQL",
          "FieldDisplayName": "JQL",
          "Type": "string",
          "Required": false
        },
        {
          "FieldName": "EstimateField",
          "FieldDisplayName": "Estimate Field (Leave blank to use TimeEstimate)",
          "Type": "string",
          "Required": false
        }
      ]
    }`
	return utils.ByteToMap([]byte(template)).(map[string]interface{})
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
	tickets, _ := c.getTicketsFromJQL("Sprint  in (" + sprint + ")")
	return tickets
}

// ValidateConfig ...
func (c *JIRAConnection) ValidateConfig() error {
	searchOptions := jira.SearchOptions{MaxResults: 1}

	_, _, err := c.client.Issue.Search(c.config.JQL, &searchOptions)
	return err
}

func (c *JIRAConnection) getTicketsFromJQL(extraJQL string) (ticketsSerialized []serializers.Task, err error) {
	searchOptions := jira.SearchOptions{MaxResults: 50000}

	jql := extraJQL + " AND " + c.config.JQL

	if extraJQL == "" {
		jql = c.config.JQL
	}

	// ToDo: Use pagination
	tickets, _, err := c.client.Issue.Search(jql, &searchOptions)
	if err != nil {
		return nil, err
	}

	for _, ticket := range tickets {
		var estimate *float64
		if c.config.EstimateField != "" {
			estimates := ticket.Fields.Unknowns[c.config.EstimateField]

			switch estimates.(type) {
			case string:
				estimateFromString, err := strconv.ParseFloat(estimates.(string), 64)
				if err.(error) == nil {
					estimate = &estimateFromString
				}
			case int:
				estimateFromInt := float64(estimates.(int))
				estimate = &estimateFromInt
			case float64:
				estimateFromFloat := estimates.(float64)
				estimate = &estimateFromFloat
			}
		} else {
			timeEstimate := float64(ticket.Fields.TimeOriginalEstimate) / 3600
			estimate = &timeEstimate
		}

		ticketsSerialized = append(ticketsSerialized, serializers.Task{
			ID:        ticket.Key,
			ProjectID: ticket.Fields.Project.ID,
			Summary:   ticket.Fields.Summary,
			Type:      ticket.Fields.Type.Name,
			Priority:  ticket.Fields.Priority.Name,
			Estimate:  estimate,
			Assignee:  ticket.Fields.Assignee.DisplayName,
			Status:    ticket.Fields.Status.Name,
		})
	}
	return ticketsSerialized,nil
}
