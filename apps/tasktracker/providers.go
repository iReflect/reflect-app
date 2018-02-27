package tasktracker

import (
	"encoding/json"

	"errors"
	"github.com/blaskovicz/go-cryptkeeper"
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
	var configList []map[string]interface{}
	if err = json.Unmarshal(decrypted, &configList); err != nil {
		return nil, err
	}

	cryptkeeper.SetCryptKey(utils.EncryptionKey())
	var providerData map[string]interface{}
	var credentials map[string]interface{}

	for _, taskProviderConfig := range configList {
		providerData = taskProviderConfig["data"].(map[string]interface{})
		credentials = providerData["credentials"].(map[string]interface{})
		if val, ok := credentials["password"]; ok {
			credentials["password"], err = cryptkeeper.Encrypt(val.(string))

			if err != nil {
				return nil, err
			}
		}
		if val, ok := credentials["apiToken"]; ok {
			credentials["apiToken"], err = cryptkeeper.Encrypt(val.(string))
			if err != nil {
				return nil, err
			}
		}
	}
	return json.Marshal(configList)
}

// DecryptTaskProviders ...
func DecryptTaskProviders(encrypted []byte) (decrypted []byte, err error) {
	var configList []map[string]interface{}
	if err = json.Unmarshal(encrypted, &configList); err != nil {
		return nil, err
	}

	cryptkeeper.SetCryptKey(utils.EncryptionKey())
	var providerData map[string]interface{}
	var credentials map[string]interface{}
	for _, taskProviderConfig := range configList {
		providerData = taskProviderConfig["data"].(map[string]interface{})
		credentials = providerData["credentials"].(map[string]interface{})
		if val, ok := credentials["password"]; ok {
			credentials["password"], err = cryptkeeper.Decrypt(val.(string))
			if err != nil {
				return nil, err
			}
		}
		if val, ok := credentials["apiToken"]; ok {
			credentials["apiToken"], err = cryptkeeper.Decrypt(val.(string))
			if err != nil {
				return nil, err
			}
		}
	}

	return json.Marshal(configList)
}

// ValidateCredentials ...
func ValidateCredentials(credentials map[string]interface{}) (err error) {
	authType, _ := credentials["type"]
	err = errors.New("invalid credentials")
	if authType == "basicAuth" {
		if _, ok := credentials["password"]; !ok {
			return err
		}
	} else if authType == "apiToken" {
		if _, ok := credentials["apiToken"]; !ok {
			return err
		}
	} else {
		return err
	}
	return nil
}

// GetTaskList ...
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

// GetSprintTaskList ...
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

// GetConnections ...
func GetConnections(config []byte) (connections []Connection, err error) {
	var configList []interface{}
	if err = json.Unmarshal(config, &configList); err != nil {
		return nil, err
	}

	var data map[string]interface{}
	var name string
	var connection Connection
	for _, tpConfig := range configList {
		tp := tpConfig.(map[string]interface{})
		data = tp["data"].(map[string]interface{})
		name = tp["type"].(string)
		connection = GetTaskProvider(name).New(data)
		if connection != nil {
			connections = append(connections, connection)
		}
	}

	return connections, nil
}

// ValidateConfigs ...
func ValidateConfigs(taskProviderConfigList []map[string]interface{}) (err error) {
	for _, taskProviderConfig := range taskProviderConfigList {
		taskProvider := GetTaskProvider(taskProviderConfig["type"].(string))
		taskProviderConnection := taskProvider.New(taskProviderConfig["data"])
		if err = taskProviderConnection.ValidateConfig(); err != nil {
			return err
		}
	}
	return nil
}
