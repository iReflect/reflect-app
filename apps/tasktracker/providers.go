package tasktracker

import (
	"encoding/json"

	"errors"
	"strings"

	"github.com/blaskovicz/go-cryptkeeper"
	"github.com/iReflect/reflect-app/apps/tasktracker/serializers"
	"github.com/iReflect/reflect-app/config"
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
	GetTaskList(ticketKeys []string) []serializers.Task
	GetTask(ticketKey string) (*serializers.Task, error)
	GetTaskUrl(ticketKey string) string
	GetSprint(sprintID string) *serializers.Sprint
	GetSprintTaskList(sprint serializers.Sprint) []serializers.Task
	ValidateConfig() error
}

// TaskProviders ...
var TaskProviders = make(map[string]TaskProvider)

// TaskTypes ...
var TaskTypes = []string{"FeatureTypes", "TaskTypes", "BugTypes"}

// DoneStatus ...
const DoneStatus = "DoneStatus"

// StatusTypes ...
var StatusTypes = []string{DoneStatus}

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

	cryptkeeper.SetCryptKey([]byte(config.GetConfig().Server.EncryptionKey))
	var providerData map[string]interface{}
	var credentials map[string]interface{}

	for _, taskProviderConfig := range configList {
		providerData = taskProviderConfig["data"].(map[string]interface{})
		credentials = providerData["credentials"].(map[string]interface{})
		if val, ok := credentials["password"]; ok {
			credentials["password"], err = cryptkeeper.Encrypt(val.(string))

			if err != nil {
				utils.LogToSentry(err)
				return nil, err
			}
		}
		if val, ok := credentials["apiToken"]; ok {
			credentials["apiToken"], err = cryptkeeper.Encrypt(val.(string))
			if err != nil {
				utils.LogToSentry(err)
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

	cryptkeeper.SetCryptKey([]byte(config.GetConfig().Server.EncryptionKey))
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
func GetTaskList(config []byte, taskKeys []string) (tasks []serializers.Task, err error) {
	if len(taskKeys) == 0 {
		return tasks, nil
	}

	connection := GetConnection(config)
	if connection == nil {
		return nil, errors.New("task_list: invalid connection config")
	}

	tasks = append(tasks, connection.GetTaskList(taskKeys)...)
	return tasks, nil
}

// GetTaskDetails ...
func GetTaskDetails(config []byte, taskKey string) (*serializers.Task, error) {

	connection := GetConnection(config)
	if connection == nil {
		return nil, errors.New("task_details: invalid connection config")
	}

	return connection.GetTask(taskKey)
}

// GetSprintTaskList ...
func GetSprintTaskList(config []byte, sprint serializers.Sprint) (tasks []serializers.Task, err error) {
	connection := GetConnection(config)
	if connection == nil {
		return nil, errors.New("sprint_task_list: invalid connection config")
	}

	tasks = append(tasks, connection.GetSprintTaskList(sprint)...)
	return tasks, nil
}

// GetConnection ...
func GetConnection(config []byte) Connection {
	var configList []interface{}
	if err := json.Unmarshal(config, &configList); err != nil {
		return nil
	}
	var data map[string]interface{}
	var name string
	if len(configList) == 0 {
		return nil
	}
	tpConfig := configList[0]
	tp := tpConfig.(map[string]interface{})
	data = tp["data"].(map[string]interface{})
	name = tp["type"].(string)
	return GetTaskProvider(name).New(data)
}

// GetTaskTypeMappings ...
func GetTaskTypeMappings(config []byte) (map[string][]string, error) {
	var configList []interface{}
	types := make(map[string][]string)

	if err := json.Unmarshal(config, &configList); err != nil {
		return nil, err
	}

	var data map[string]interface{}
	for _, tpConfig := range configList {
		tp := tpConfig.(map[string]interface{})
		data = tp["data"].(map[string]interface{})

		for _, taskType := range TaskTypes {
			typeUpper, ok := data[taskType].(string)
			if !ok {
				return nil, errors.New("failed to read from retrospective config")
			}
			// remove extra spaces from the task tracker type mapping values
			taskTypeValueList := strings.Split(strings.ToLower(typeUpper), ",")
			for index, value := range taskTypeValueList {
				taskTypeValueList[index] = strings.TrimSpace(value)
			}
			types[taskType] = taskTypeValueList
		}
	}

	return types, nil
}

// GetStatusMapping ...
func GetStatusMapping(config []byte) (map[string][]string, error) {
	var configList []interface{}
	statusType := make(map[string][]string)

	if err := json.Unmarshal(config, &configList); err != nil {
		return nil, err
	}

	var data map[string]interface{}
	for _, tpConfig := range configList {
		tp := tpConfig.(map[string]interface{})
		data = tp["data"].(map[string]interface{})

		for _, status := range StatusTypes {
			statusUpper, ok := data[status].(string)
			if ok {
				// remove extra spaces from the completed task status mapping values
				statusTypeList := strings.Split(strings.ToLower(statusUpper), ",")
				for index, status := range statusTypeList {
					statusTypeList[index] = strings.TrimSpace(status)
				}
				statusType[status] = statusTypeList
			} else {
				statusType[status] = []string{}
			}
		}
	}
	return statusType, nil
}

// ValidateConfigs ...
func ValidateConfigs(taskProviderConfigList []map[string]interface{}) (err error) {
	for _, taskProviderConfig := range taskProviderConfigList {
		taskProvider := GetTaskProvider(taskProviderConfig["type"].(string))
		taskProviderConnection := taskProvider.New(taskProviderConfig["data"])
		if err = taskProviderConnection.ValidateConfig(); err != nil {
			utils.LogToSentry(err)
			return errors.New("failed to validate config for " + taskProviderConfig["type"].(string))
		}
	}
	return nil
}
