package providers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/andygrunwald/go-jira"

	"encoding/json"
	"errors"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	"github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
	"io/ioutil"
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
	jiraConfig, err := getConfigObject(config)

	if err != nil {
		return nil
	}

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

// getConfigObject ...
func getConfigObject(config interface{}) (JIRAConfig, error) {
	var c JIRAConfig

	switch config.(type) {
	case []byte:
		c = JIRAConfig{}
		err := json.Unmarshal(config.([]byte), &c)
		if err != nil {
			return c, err
		}
	case map[string]interface{}:
		c = JIRAConfig{}

		jsonConfig, err := json.Marshal(config)
		if err != nil {
			return c, err
		}

		err = json.Unmarshal(jsonConfig, &c)
		if err != nil {
			return c, err
		}
	case JIRAConfig:
		c = config.(JIRAConfig)
	default:
		return c, errors.New("invalid type")
	}
	return c, nil
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
func (c *JIRAConnection) GetTaskList(ticketKeys []string) []serializers.Task {
	tickets, err := c.getTicketsFromJQL("issueKey in ("+strings.Join(ticketKeys, ",")+")", true)

	if err != nil {
		utils.LogToSentry(err)
	}

	return tickets
}

// GetTask ...
func (c *JIRAConnection) GetTask(ticketKey string) (*serializers.Task, error) {
	ticket, err := c.getTicket(ticketKey)
	if err != nil {
		utils.LogToSentry(err)
		return nil, err
	}
	return ticket, nil
}

// GetSprint ...
func (c *JIRAConnection) GetSprint(sprintID string) *serializers.Sprint {
	boardIDs := strings.Split(c.config.BoardIds, ",")

	var sprints []jira.Sprint
	for _, boardID := range boardIDs {
		sprint, _, err := c.client.Board.GetAllSprints(boardID)

		if err == nil {
			sprints = append(sprints, sprint...)
		} else {
			utils.LogToSentry(err)
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

	tickets, _ := c.getTicketsFromJQL("Sprint  in ("+sprint+")", false)
	return tickets
}

// ValidateConfig ...
func (c *JIRAConnection) ValidateConfig() error {
	searchOptions := jira.SearchOptions{MaxResults: 1}

	_, _, err := c.client.Issue.Search(c.config.JQL, &searchOptions)
	return err
}

func (c *JIRAConnection) getTicketsFromJQL(extraJQL string, skipBaseJQL bool) (ticketsSerialized []serializers.Task, err error) {
	// Need to pass in validateQuery=warn like this until jira-go supports this natively
	searchOptions := jira.SearchOptions{MaxResults: 50000, ValidateQuery: "warn"}

	var jql string

	switch {
	case skipBaseJQL:
		jql = extraJQL
	case extraJQL != "" && c.config.JQL != "":
		jql = extraJQL + " AND " + c.config.JQL
	case extraJQL != "":
		jql = extraJQL
	case c.config.JQL != "":
		jql = c.config.JQL
	}

	// ToDo: Use pagination
	tickets, res, err := c.client.Issue.Search(jql, &searchOptions)
	if err != nil {
		jiraErr, _ := ioutil.ReadAll(res.Response.Body)
		utils.LogToSentry(errors.New(string(jiraErr)))
		return nil, err
	}

	return c.serializeTickets(tickets), nil
}

func (c *JIRAConnection) getTicket(ticketKey string) (ticketSerialized *serializers.Task, err error) {

	ticket, res, err := c.client.Issue.Get(ticketKey, nil)
	if err != nil {
		jiraErr, _ := ioutil.ReadAll(res.Response.Body)

		if strings.Contains(err.Error(), "Issue does not exist") {
			return nil, nil
		}
		utils.LogToSentry(fmt.Errorf("%s: %s", ticketKey, jiraErr))
		return nil, err
	}

	if ticket == nil {
		return nil, nil
	}

	return c.serializeTicket(*ticket), nil
}

func (c *JIRAConnection) serializeTickets(tickets []jira.Issue) (ticketsSerialized []serializers.Task) {
	for _, ticket := range tickets {
		ticketsSerialized = append(ticketsSerialized, *c.serializeTicket(ticket))
	}

	return ticketsSerialized
}

func (c *JIRAConnection) serializeTicket(ticket jira.Issue) *serializers.Task {
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

	return &serializers.Task{
		Key:             ticket.Key,
		TrackerUniqueID: ticket.ID,
		ProjectID:       ticket.Fields.Project.ID,
		Summary:         ticket.Fields.Summary,
		Description:     ticket.Fields.Description,
		Type:            ticket.Fields.Type.Name,
		Priority:        ticket.Fields.Priority.Name,
		Estimate:        estimate,
		Assignee:        assignee,
		Status:          ticket.Fields.Status.Name,
	}
}
