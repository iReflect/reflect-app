package tasktracker

import (
	"encoding/json"

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
	ConfigTemplate() []byte
}

// Connection ...
type Connection interface {
	GetTaskList(query string) []serializers.Task
	GetSprint(sprint string) *serializers.Sprint
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
func EncryptTaskProviders(decrypted []byte) (encrypted []byte, err error) {
	configMap, ok := utils.ByteToMap(decrypted).([]map[string]interface{})

	if !ok {
		return nil, nil
	}

	var data map[string]interface{}
	var creds map[string][]byte
	for _, tp := range configMap {
		data = tp["data"].(map[string]interface{})
		creds = data["credentials"].(map[string][]byte)
		if val, ok := creds["password"]; ok {
			creds["password"], err = utils.EncryptString(val)
			if err != nil {
				return nil, err
			}
		}
		if val, ok := creds["apiToken"]; ok {
			creds["apiToken"], err = utils.EncryptString(val)
			if err != nil {
				return nil, err
			}
		}
	}

	return json.Marshal(configMap)
}

// DecryptTaskProviders ...
func DecryptTaskProviders(encrypted []byte) (decrypted []byte, err error) {
	configMap, ok := utils.ByteToMap(encrypted).([]map[string]interface{})

	if !ok {
		return nil, nil
	}

	var data map[string]interface{}
	var creds map[string][]byte
	for _, tp := range configMap {
		data = tp["data"].(map[string]interface{})
		creds = data["credentials"].(map[string][]byte)
		if val, ok := creds["password"]; ok {
			creds["password"], err = utils.DecryptString(val)
			if err != nil {
				return nil, err
			}
		}
		if val, ok := creds["apiToken"]; ok {
			creds["apiToken"], err = utils.DecryptString(val)
			if err != nil {
				return nil, err
			}
		}
	}

	return json.Marshal(configMap)
}
