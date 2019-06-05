package providers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/andygrunwald/go-jira"

	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/iReflect/reflect-app/apps/tasktracker"
	"github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"github.com/iReflect/reflect-app/constants"
	"github.com/iReflect/reflect-app/libs/utils"
)

// SprintIDJQLKeyword ...
const SprintIDJQLKeyword = "${sprintID}"

// FromDateJQLKeyword ...
const FromDateJQLKeyword = "${fromDate}"

// ToDateJQLKeyword ...
const ToDateJQLKeyword = "${toDate}"

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

func (config JIRAConfig) GetBaseURL() string {
	// Just to make sure there are no trailing slashes in the base url, even if provided by the user.
	return strings.Trim(config.BaseURL, "/")
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

	client, err := jira.NewClient(nil, jiraConfig.GetBaseURL())

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
	configMap = map[string]interface{}{
		"Type":               TaskProviderJira,
		"DisplayTitle":       "JIRA",
		"SupportedAuthTypes": []string{"basicAuth"},
		"Fields": []map[string]interface{}{
			{
				"FieldName":        "BaseURL",
				"FieldDisplayName": "Base URL of the project. eg. 'https://ireflect.atlassian.net'",
				"Type":             "string",
				"Required":         true,
			},
			{
				"FieldName":        "BoardIds",
				"FieldDisplayName": "Board IDs (Comma Separated)",
				"Type":             "string",
				"Required":         true,
			},
			{
				"FieldName": "JQL",
				"FieldDisplayName": fmt.Sprintf("JQL. eg. priority in (Blocker, Critical) AND status was \"Open\" During (%s, %s)",
					FromDateJQLKeyword, ToDateJQLKeyword),
				"Type":     "string",
				"Required": false,
				"Hint": fmt.Sprintf("<i>You can use the following parameters in your custom JQL, which will be replaced with "+
					"their actual values at the time of the sprint sync.<br><strong>Sprint ID</strong>: %s, "+
					"<strong>\"From\" Date</strong>: %s, <strong>\"To\" Date</strong>: %s </i>",
					SprintIDJQLKeyword, FromDateJQLKeyword, ToDateJQLKeyword),
			},
			{
				"FieldName":        "EstimateField",
				"FieldDisplayName": "Estimate Field (Leave blank to use TimeEstimate)",
				"Type":             "string",
				"Required":         false,
			},
		},
	}
	return configMap
}

// SanitizeTimeLogs ...
func (m *JIRAConnection) SanitizeTimeLogs(timeLogKeys []string) map[string]string {
	// This method is currently not used
	return nil
}

// GetTaskUrl ...
func (c *JIRAConnection) GetTaskUrl(ticketKey string) string {
	return fmt.Sprintf("%v/browse/%v", c.config.GetBaseURL(), ticketKey)
}

// GetTaskList ...
func (c *JIRAConnection) GetTaskList(ticketKeys []string) []serializers.Task {
	tickets, err := c.getTicketsFromJQL(fmt.Sprintf("issueKey in (%s)", strings.Join(ticketKeys, ",")), true, nil)

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
	// Prepare http request for the sprint Get
	req, err := c.client.NewRequest("GET", fmt.Sprintf("rest/agile/1.0/sprint/%s", sprintID), nil)
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}

	var sprint jira.Sprint
	// Call the Sprint Get API
	resp, err := c.client.Do(req, &sprint)
	if err != nil {
		utils.LogToSentry(jira.NewJiraError(resp, err))
		return nil
	}

	return &serializers.Sprint{
		ID:       sprintID,
		BoardID:  strconv.Itoa(sprint.OriginBoardID),
		Name:     sprint.Name,
		FromDate: sprint.StartDate,
		ToDate:   sprint.EndDate,
	}
}

// GetSprintTaskList ...
func (c *JIRAConnection) GetSprintTaskList(sprint serializers.Sprint) []serializers.Task {
	var extraJQL string
	if sprint.ID != "" {
		extraJQL = fmt.Sprintf("Sprint in (%s)", sprint.ID)
	}
	tickets, _ := c.getTicketsFromJQL(extraJQL, false, &sprint)
	return tickets
}

// ValidateConfig ...
func (c *JIRAConnection) ValidateConfig() error {
	searchOptions := jira.SearchOptions{MaxResults: 1}

	_, _, err := c.client.Issue.Search("", &searchOptions)
	return err
}

func (c *JIRAConnection) getTicketsFromJQL(extraJQL string, skipBaseJQL bool, sprint *serializers.Sprint) (ticketsSerialized []serializers.Task, err error) {
	// Need to pass in validateQuery=warn like this until jira-go supports this natively
	searchOptions := jira.SearchOptions{MaxResults: 50000, ValidateQuery: "warn"}

	jql := ""
	if !skipBaseJQL && c.config.JQL != "" {
		jql = c.sanitizeJQL(sprint)
		if extraJQL != "" {
			jql = extraJQL + " AND " + jql
		}
	} else {
		jql = extraJQL
	}
	jql = strings.Trim(jql, " ")
	// If the sanitized JQL is empty, then there is no need to get tickets from the JIRA Board
	if jql == "" {
		return nil, nil
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

	ticket, resp, err := c.client.Issue.Get(ticketKey, nil)
	if err != nil {

		if resp.StatusCode == 404 {
			return nil, nil
		}

		jiraErr, _ := ioutil.ReadAll(resp.Response.Body)
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

	serializedTask := serializers.Task{
		Key:             ticket.Key,
		TrackerUniqueID: ticket.ID,
		Estimate:        estimate,
		ProjectID:       ticket.Fields.Project.ID,
		Summary:         ticket.Fields.Summary,
		Description:     ticket.Fields.Description,
		Type:            ticket.Fields.Type.Name,
	}

	if ticket.Fields.Assignee != nil {
		serializedTask.Assignee = ticket.Fields.Assignee.DisplayName
	}
	if ticket.Fields.Status != nil {
		serializedTask.Status = ticket.Fields.Status.Name
	}
	if ticket.Fields.Priority != nil {
		serializedTask.Priority = ticket.Fields.Priority.Name
	}

	return &serializedTask
}

// sanitizeJQL replaces the parameters in the JQL with their respective values
func (c *JIRAConnection) sanitizeJQL(sprint *serializers.Sprint) string {
	if sprint == nil {
		return ""
	}
	fromDate, toDate := "", ""
	if sprint.FromDate != nil {
		fromDate = sprint.FromDate.Format(constants.CustomDateFormat)
	}
	if sprint.ToDate != nil {
		// Adding 1 day to include the to date in the calculations
		toDate = sprint.ToDate.AddDate(0, 0, 1).Format(constants.CustomDateFormat)
	}
	return strings.NewReplacer(SprintIDJQLKeyword, sprint.ID, FromDateJQLKeyword, fromDate, ToDateJQLKeyword, toDate).Replace(c.config.JQL)
}
