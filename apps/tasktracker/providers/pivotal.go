package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iReflect/go-pivotaltracker/v5/pivotal"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	"github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
	"strconv"
	"strings"
)

// PivotalTaskProvider ...
type PivotalTaskProvider struct {
}

// PivotalConnection ...
type PivotalConnection struct {
	config PivotalConfig
	client *pivotal.Client
}

// PivotalConfig ...
type PivotalConfig struct {
	Credentials tasktracker.Credentials `json:"Credentials"`
	ProjectID   string                  `json:"ProjectID"`
}

// TaskProviderPivotal ...
const TaskProviderPivotal = "pivotal"

func init() {
	provider := &PivotalTaskProvider{}
	tasktracker.RegisterTaskProvider(TaskProviderPivotal, provider)
}

// New ...
func (p *PivotalTaskProvider) New(config interface{}) tasktracker.Connection {
	var pivotalConfig PivotalConfig
	var client *pivotal.Client
	pivotalConfig, err := getPivotalConfigObject(config)

	if err != nil {
		return nil
	}

	switch pivotalConfig.Credentials.Type {
	case "apiToken":
		client = pivotal.NewClient(pivotalConfig.Credentials.APIToken)
	default:
		return nil
	}
	return &PivotalConnection{config: pivotalConfig, client: client}
}

// getPivotalConfigObject ...
func getPivotalConfigObject(config interface{}) (PivotalConfig, error) {
	var c PivotalConfig

	switch config.(type) {
	case []byte:
		c = PivotalConfig{}
		err := json.Unmarshal(config.([]byte), &c)
		if err != nil {
			return c, err
		}
	case map[string]interface{}:
		c = PivotalConfig{}

		jsonConfig, err := json.Marshal(config)
		if err != nil {
			return c, err
		}

		err = json.Unmarshal(jsonConfig, &c)
		if err != nil {
			return c, err
		}
	case PivotalConfig:
		c = config.(PivotalConfig)
	default:
		return c, errors.New("invalid type")
	}
	return c, nil
}

// ConfigTemplate ...
func (p *PivotalTaskProvider) ConfigTemplate() (configMap map[string]interface{}) {
	configMap = map[string]interface{}{
		"Type":               TaskProviderPivotal,
		"DisplayTitle":       "Pivotal Tracker",
		"SupportedAuthTypes": []string{"apiToken"},
		"Fields": []map[string]interface{}{
			{
				"FieldName":        "ProjectID",
				"FieldDisplayName": "Enter the Project ID for your PT Project. e.g. 1234567",
				"Type":             "number",
				"Required":         true,
			},
		},
	}
	return configMap
}

// getProjectUsers ...
func (c *PivotalConnection) getProjectUsers() ([]pivotal.Person, error) {
	projectID, err := strconv.Atoi(c.config.ProjectID)
	if err != nil {
		return nil, err
	}

	projectMemberships, _, err := c.client.Memberships.List(projectID)
	var projectUsers []pivotal.Person
	if err != nil {
		return nil, err
	}

	for _, projectMembership := range projectMemberships {
		projectUsers = append(projectUsers, projectMembership.Person)
	}

	return projectUsers, err
}

// getUserIDNameMap ...
func (c *PivotalConnection) getUserIDNameMap() map[int]string {
	projectUsers, err := c.getProjectUsers()

	if err != nil {
		return nil
	}

	userIDNameMap := make(map[int]string)
	for _, user := range projectUsers {
		userIDNameMap[user.Id] = user.Name
	}

	return userIDNameMap
}

// serializeTickets ...
func (c *PivotalConnection) serializeTickets(tickets []*pivotal.Story, userIDNameMap map[int]string) (ticketsSerialized []serializers.Task) {

	for _, ticket := range tickets {
		ticketsSerialized = append(ticketsSerialized, *c.serializeTicket(ticket, userIDNameMap))
	}

	return ticketsSerialized
}

// serializeTicket ...
func (c *PivotalConnection) serializeTicket(ticket *pivotal.Story, userIDNameMap map[int]string) *serializers.Task {
	ticketID := strconv.Itoa(ticket.Id)
	task := &serializers.Task{
		Key:             ticketID,
		TrackerUniqueID: ticketID,
		Summary:         ticket.Name,
		Description:     ticket.Description,
		Type:            ticket.Type,
		Estimate:        ticket.Estimate,
		Status:          ticket.State,
		ProjectID:       c.config.ProjectID,
		Priority:        "",
	}

	// Set Assignee of a task, since PT has owners (multiple) so we can set it to a comma separated list of Owner names
	if userIDNameMap != nil {
		var owners []string
		for _, ownerId := range ticket.OwnerIds {
			if ownerName, isPresent := userIDNameMap[ownerId]; isPresent {
				owners = append(owners, ownerName)
			}
		}
		task.Assignee = strings.Join(owners, ", ")
	}
	return task
}

// GetTaskList ...
func (c *PivotalConnection) GetTaskList(ticketKeys []string) []serializers.Task {
	if len(ticketKeys) == 0 {
		return nil
	}

	filterQuery := fmt.Sprintf("id:%s", strings.Join(ticketKeys, ","))

	projectID, err := strconv.Atoi(c.config.ProjectID)
	if err != nil {
		return nil
	}

	stories, err := c.client.Stories.List(projectID, filterQuery)
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}

	return c.serializeTickets(stories, c.getUserIDNameMap())
}

// GetTask ...
func (c *PivotalConnection) GetTask(ticketKey string) (*serializers.Task, error) {
	ticketID, err := strconv.Atoi(ticketKey)
	// Since the ticket ids are always number in Pivotal, if a value is not
	if err != nil {
		utils.LogToSentry(err)
		return nil, nil
	}
	projectID, err := strconv.Atoi(c.config.ProjectID)
	if err != nil {
		return nil, err
	}

	story, _, err := c.client.Stories.Get(projectID, ticketID)
	if err != nil {
		utils.LogToSentry(err)
		return nil, err
	}

	return c.serializeTicket(story, c.getUserIDNameMap()), nil
}

// GetSprint ...
func (c *PivotalConnection) GetSprint(sprintID string) *serializers.Sprint {
	iterationNumber, err := strconv.Atoi(sprintID)
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}
	projectID, err := strconv.Atoi(c.config.ProjectID)
	if err != nil {
		return nil
	}
	iteration, _, err := c.client.Iterations.Get(projectID, iterationNumber)
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}
	return &serializers.Sprint{
		ID:       sprintID,
		BoardID:  "",
		Name:     fmt.Sprintf("Iteration-%v", strconv.Itoa(iteration.Number)),
		FromDate: iteration.Start,
		ToDate:   iteration.Finish,
	}
}

// GetSprintTaskList ...
func (c *PivotalConnection) GetSprintTaskList(sprint serializers.Sprint) []serializers.Task {
	iterationNumber, err := strconv.Atoi(sprint.ID)
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}
	projectID, err := strconv.Atoi(c.config.ProjectID)
	if err != nil {
		return nil
	}
	iteration, _, err := c.client.Iterations.Get(projectID, iterationNumber)
	if err != nil {
		utils.LogToSentry(err)
		return nil
	}

	var storyIDs []string
	for _, story := range iteration.Stories {
		storyIDs = append(storyIDs, strconv.Itoa(story.Id))
	}

	return c.GetTaskList(storyIDs)
}

// ValidateConfig validates if the provided API Token and ProjectID are correct
func (c *PivotalConnection) ValidateConfig() error {
	projectID, err := strconv.Atoi(c.config.ProjectID)
	if err != nil {
		return nil
	}
	_, _, err = c.client.Projects.Get(projectID)
	return err
}
