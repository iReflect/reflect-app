package providers

import (
	"encoding/json"
	"errors"
	"github.com/iReflect/reflect-app/apps/tasktracker"
	"github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"gopkg.in/salsita/go-pivotaltracker.v1/v5/pivotal"
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
	Credentials  tasktracker.Credentials `json:"Credentials"`
	ProjectID    int                     `json:"ProjectID"`
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

// GetTaskList ...
func (c *PivotalConnection) GetTaskList(ticketKeys []string) []serializers.Task {
	return nil
}

// GetTask ...
func (c *PivotalConnection) GetTask(ticketKey string) (*serializers.Task, error) {
	return nil, nil
}

// GetSprint ...
func (c *PivotalConnection) GetSprint(sprintID string) *serializers.Sprint {
	return nil
}

// GetSprintTaskList ...
func (c *PivotalConnection) GetSprintTaskList(sprint serializers.Sprint) []serializers.Task {
	return nil
}

// ValidateConfig validates if the provided API Token and ProjectID are correct
func (c *PivotalConnection) ValidateConfig() error {
	_, _, err := c.client.Projects.Get(c.config.ProjectID)
	return err
}
