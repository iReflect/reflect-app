package tasktracker

import (
	"encoding/json"

	"errors"
	"github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"github.com/iReflect/reflect-app/libs/utils"
)

// Credentials ...
type Credentials struct {
	Type     string `json:"Type"`
	Username string `json:"Username"`
	Password string `json:"Password"`
	APIToken string `json:"APIToken"`
}

// TaskProvider ...
type TaskProvider interface {
	New(config interface{}) Connection
	ConfigTemplate() map[string]interface{}
}

// Connection ...
type Connection interface {
	GetTaskList(ticketIDs []string) []serializers.Task
	GetSprint(sprintID string) *serializers.Sprint
	GetSprintTaskList(sprint string) []serializers.Task
	ValidateConfig() error
}

// TaskProviders ...
var TaskProviders = make(map[string]TaskProvider)

// RegisterTaskProvider ...
func RegisterTaskProvider(name string, newProvider TaskProvider) {
	TaskProviders[name] = newProvider
}

// GetTaskProvider ...
func GetTaskProvider(name string) TaskProvider {
	provider, ok := TaskProviders[name]
	if ok {
		return provider
	}
	return nil
}

// EncryptTaskProviders ...
//ToDo: Generalize these methods
func EncryptTaskProviders(decrypted []byte) (encrypted []byte, err error) {
	configMap, ok := utils.ByteToMap(decrypted).([]interface{})

	if !ok {
		return nil, errors.New("JSON Error")
	}

	var data map[string]interface{}
	var creds map[string]interface{}
	for _, tpConfig := range configMap {
		tp := tpConfig.(map[string]interface{})
		data = tp["data"].(map[string]interface{})
		creds = data["credentials"].(map[string]interface{})
		if val, ok := creds["password"]; ok {
			creds["password"], err = utils.EncryptString(val.([]byte))
			if err != nil {
				return nil, err
			}
		}
		if val, ok := creds["apiToken"]; ok {
			creds["apiToken"], err = utils.EncryptString(val.([]byte))
			if err != nil {
				return nil, err
			}
		}
	}

	return json.Marshal(configMap)
}

// DecryptTaskProviders ...
func DecryptTaskProviders(encrypted []byte) (decrypted []byte, err error) {
	configMap, ok := utils.ByteToMap(encrypted).([]interface{})

	if !ok {
		return nil, errors.New("JSON Error")
	}

	var data map[string]interface{}
	var creds map[string]interface{}
	for _, tpConfig := range configMap {
		tp := tpConfig.(map[string]interface{})
		data = tp["data"].(map[string]interface{})
		creds = data["credentials"].(map[string]interface{})
		if val, ok := creds["password"]; ok {
			creds["password"], err = utils.DecryptString(val.([]byte))
			if err != nil {
				return nil, err
			}
		}
		if val, ok := creds["apiToken"]; ok {
			creds["apiToken"], err = utils.DecryptString(val.([]byte))
			if err != nil {
				return nil, err
			}
		}
	}

	return json.Marshal(configMap)
}

func GetTaskList(config []byte, taskIDs []string) (tasks []serializers.Task, err error) {
	connections, err := GetConnections(config)
	if err != nil {
		return nil, err
	}

	for _, connection := range connections {
		tasks = append(tasks, connection.GetTaskList(taskIDs)...)
	}
	return tasks, nil
}

func GetSprintTaskList(config []byte, sprintIDs string) (tasks []serializers.Task, err error) {
	connections, err := GetConnections(config)
	if err != nil {
		return nil, err
	}

	for _, connection := range connections {
		tasks = append(tasks, connection.GetSprintTaskList(sprintIDs)...)
	}
	return tasks, nil
}

func GetConnections(config []byte) (connections []Connection, err error) {
	configMap, ok := utils.ByteToMap(config).([]interface{})

	if !ok {
		return nil, errors.New("JSON Error")
	}

	var data map[string]interface{}
	var name string
	var connection Connection
	for _, tpConfig := range configMap {
		tp := tpConfig.(map[string]interface{})
		data = tp["data"].(map[string]interface{})
		name = tp["name"].(string)
		connection = GetTaskProvider(name).New(data)
		if connection != nil {
			connections = append(connections, connection)
		}
	}

	return connections, nil
}
