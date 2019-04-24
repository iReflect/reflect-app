package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/iReflect/reflect-app/constants"

	"github.com/iReflect/reflect-app/apps/timetracker"
	"github.com/iReflect/reflect-app/apps/timetracker/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
)

// JiraTimeProvider ...
type JiraTimeProvider struct {
}

// JIRAConnection ...
type JIRAConnection struct {
	config JIRAConfig
	client *jira.Client
}

// JIRAConfig ...
type JIRAConfig struct {
	Credentials   Credentials `json:"Credentials"`
	BaseURL       string      `json:"BaseURL"`
	BoardIds      string      `json:"BoardIds"`
	JQL           string      `json:"JQL"`
	EstimateField string      `json:"EstimateField"`
}

// Credentials ...
type Credentials struct {
	Type     string `json:"Type"`
	Username string `json:"Username"`
	Password string `json:"Password"`
}

// JiraTimeResult ...
type JiraTimeResult struct {
	Project string  `json:"Project"`
	TaskID  string  `json:"TaskID"`
	Hours   float64 `json:"Hours"`
}

// TimeProviderJira ...
const (
	TimeProviderJira            = "jira"
	TimeProviderJiraDisplayName = "JIRA"
)

func init() {
	provider := &JiraTimeProvider{}
	timetracker.RegisterTimeProvider(TimeProviderJira, provider)
	timetracker.RegisterTimeProviderDisplayName(TimeProviderJira, TimeProviderJiraDisplayName)
}

// New ...
func (jiraConnection *JiraTimeProvider) New(config interface{}) timetracker.Connection {
	var jiraConfig JIRAConfig
	jiraConfig, err := getJIRAConfigObject(config)

	if err != nil {
		return nil
	}
	client, err := jira.NewClient(nil, jiraConfig.GetBaseURL())
	if err != nil {
		return nil
	}

	switch jiraConfig.Credentials.Type {
	case "basicAuth":
		{
			client.Authentication.SetBasicAuth(jiraConfig.Credentials.Username, jiraConfig.Credentials.Password)
		}
	default:
		return nil
	}

	return &JIRAConnection{config: jiraConfig, client: client}
}

// GetBaseURL ...
func (config JIRAConfig) GetBaseURL() string {
	// Just to make sure there are no trailing slashes in the base url, even if provided by the user.
	return strings.Trim(config.BaseURL, "/")
}

// getJIRAConfigObject ...
func getJIRAConfigObject(config interface{}) (JIRAConfig, error) {
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

// GetProjectTimeLogs ...
func (jiraConnection *JIRAConnection) GetProjectTimeLogs(project string, startTime time.Time, endTime time.Time) []serializers.TimeLog {

	var timeLogs []serializers.TimeLog
	searchOptions := jira.SearchOptions{MaxResults: 50000, Fields: []string{"worklog", "project"}, ValidateQuery: "warn"}

	// constructed JQL for a fetching ticket which had worklogs for a particular time period and project.
	jql := fmt.Sprintf("project='%s' AND worklogDate>=%s AND worklogDate<=%s", project, startTime.Format(constants.CustomDateFormat), endTime.Format(constants.CustomDateFormat))

	// added base JQL if available.
	if jiraConnection.config.JQL != "" {
		jql = fmt.Sprintf("%s AND %s", jql, jiraConnection.config.JQL)
	}

	tickets, res, err := jiraConnection.client.Issue.Search(jql, &searchOptions)
	if err != nil || res.StatusCode > 299 {
		jiraErr, _ := ioutil.ReadAll(res.Response.Body)
		utils.LogToSentry(errors.New(string(jiraErr)))
		return timeLogs
	}
	for _, ticket := range tickets {
		emailTimeMap := make(map[string]uint)
		// creating email time map for a particular ticket. i.e { "x@email.com": 100 }
		for _, worklog := range ticket.Fields.Worklog.Worklogs {
			worklogStartTime := time.Time(*worklog.Started)
			// check for worklogs which are in b/w start and end date of the sprint.
			if worklogStartTime.After(startTime) && worklogStartTime.Before(endTime) {
				emailTimeMap[worklog.Author.EmailAddress] += uint(worklog.TimeSpentSeconds)
			}
		}
		// converting map to timelogs.
		for email, timespent := range emailTimeMap {
			timeLogs = append(timeLogs, serializers.TimeLog{
				Project: ticket.Fields.Project.Name,
				TaskKey: ticket.Key,
				Minutes: timespent / 60, // converting to minutes.
				Logger:  "JIRA",
				Email:   email,
			})
		}
	}

	log.Println("Result : ", timeLogs)

	return timeLogs
}
