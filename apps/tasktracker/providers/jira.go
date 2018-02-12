package providers

import (
	"strconv"
	"strings"

	"github.com/andygrunwald/go-jira"

	"encoding/json"
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
	Credentials   tasktracker.Credentials `json:"Credentials"`
	BaseURL       string                  `json:"BaseURL"`
	BoardIds      string                  `json:"BoardIds"`
	JQL           string                  `json:"JQL"`
	EstimateField string                  `json:"EstimateField"`
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
	var jiraConfig JIRAConfig
	config = p.getConfigObject(config)

	if config == nil {
		return nil
	}

	jiraConfig = config.(JIRAConfig)
	client, err := jira.NewClient(nil, jiraConfig.BaseURL)

	if err != nil {
		return nil
	}
	switch jiraConfig.Credentials.Type {
	case "basicAuth":
		client.Authentication.SetBasicAuth(jiraConfig.Credentials.Username, jiraConfig.Credentials.Password)
	default:
		return nil
	}
	return &JIRAConnection{config: jiraConfig, client: client}
}

// GetConfigObject ...
func (p *JIRATaskProvider) getConfigObject(config interface{}) interface{} {
	var c JIRAConfig

	switch config.(type) {
	case []byte:
		c = JIRAConfig{}
		err := json.Unmarshal(config.([]byte), &c)
		if err != nil {
			return nil
		}
	case map[string]interface{}:
		c = JIRAConfig{}

		jsonConfig, err := json.Marshal(config)
		if err != nil {
			return nil
		}

		err = json.Unmarshal(jsonConfig, &c)
		if err != nil {
			return nil
		}
	case JIRAConfig:
		c = config.(JIRAConfig)
	default:
		return nil
	}
	return c
}

// ConfigTemplate ...
func (p *JIRATaskProvider) ConfigTemplate() (configMap map[string]interface{}) {
	template := `{
      "Type": "jira",
      "DisplayTitle": "JIRA",
      "SupportedAuthTypes": ["basicAuth"],
      "Fields": [
        {
          "FieldName": "BaseURL",
          "FieldDisplayName": "Base URL of the project. eg. 'https://ireflect.atlassian.net'",
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
	json.Unmarshal([]byte(template), &configMap)
	return configMap
}

// GetTaskList ...
func (c *JIRAConnection) GetTaskList(ticketIDs []string) []serializers.Task {
	var ticket *jira.Issue
	var tickets []jira.Issue
	for _, ticketID := range ticketIDs {
		ticket, _, _ = c.client.Issue.Get(ticketID, nil)
		if ticket != nil {
			tickets = append(tickets, *ticket)
		}
	}

	return c.serializeTickets(tickets)
}

// GetSprint ...
func (c *JIRAConnection) GetSprint(sprintID string) *serializers.Sprint {
	boardIDs := strings.Split(c.config.BoardIds, ",")

	var sprints []jira.Sprint
	for _, boardID := range boardIDs {
		sprint, _, err := c.client.Board.GetAllSprints(boardID)

		if err != nil {
			sprints = append(sprints, sprint...)
		}
	}

	for _, sprint := range sprints {
		if strconv.Itoa(sprint.ID) == sprintID {
			return &serializers.Sprint{
				ID:       sprintID,
				BoardID:  strconv.Itoa(sprint.OriginBoardID),
				Name:     sprint.Name,
				FromDate: sprint.StartDate,
				ToDate:   sprint.EndDate,
			}
		}
	}

	return nil
}

// GetSprintTaskList ...
func (c *JIRAConnection) GetSprintTaskList(sprint string) []serializers.Task {
	if sprint == "" {
		return nil
	}

	tickets, _ := c.getTicketsFromJQL("Sprint  in (" + sprint + ")")
	return tickets
}

// ValidateConfig ...
func (c *JIRAConnection) ValidateConfig() error {
	searchOptions := jira.SearchOptions{MaxResults: 1}

	// Todo Verify
	_, _, err := c.client.Issue.Search(c.config.JQL, &searchOptions)
	return err
}

func (c *JIRAConnection) getTicketsFromJQL(extraJQL string) (ticketsSerialized []serializers.Task, err error) {
	searchOptions := jira.SearchOptions{MaxResults: 50000}

	var jql string

	switch {
	case extraJQL != "" && c.config.JQL != "":
		jql = extraJQL + " AND " + c.config.JQL
	case extraJQL != "":
		jql = extraJQL
	case c.config.JQL != "":
		jql = c.config.JQL
	}

	// ToDo: Use pagination
	tickets, _, err := c.client.Issue.Search(jql, &searchOptions)
	if err != nil {
		return nil, err
	}

	return c.serializeTickets(tickets), nil
}

func (c *JIRAConnection) serializeTickets(tickets []jira.Issue) (ticketsSerialized []serializers.Task) {
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

		assignee := ""
		if ticket.Fields.Assignee != nil {
			assignee = ticket.Fields.Assignee.DisplayName
		}

		ticketsSerialized = append(ticketsSerialized, serializers.Task{
			ID:        ticket.Key,
			ProjectID: ticket.Fields.Project.ID,
			Summary:   ticket.Fields.Summary,
			Type:      ticket.Fields.Type.Name,
			Priority:  ticket.Fields.Priority.Name,
			Estimate:  estimate,
			Assignee:  assignee,
			Status:    ticket.Fields.Status.Name,
		})
	}

	return ticketsSerialized
}
