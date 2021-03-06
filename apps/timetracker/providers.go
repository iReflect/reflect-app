package timetracker

import (
	"encoding/json"
	"time"

	"github.com/iReflect/reflect-app/apps/timetracker/serializers"
)

// TimeProvider ...
type TimeProvider interface {
	New(config interface{}) Connection
}

// Connection ...
type Connection interface {
	GetProjectTimeLogs(project string, startTime time.Time, endTime time.Time) []serializers.TimeLog
	CleanTimeProviderConfig() interface{}
}

var timeProviders = make(map[string]TimeProvider)

// TimeProvidersDisplayNameMap ...
var TimeProvidersDisplayNameMap = make(map[string]string)

// RegisterTimeProvider ...
func RegisterTimeProvider(name string, newProvider TimeProvider) {
	timeProviders[name] = newProvider
}

// RegisterTimeProviderDisplayName ...
func RegisterTimeProviderDisplayName(name string, displayName string) {
	TimeProvidersDisplayNameMap[name] = displayName
}

// GetTimeProvider ...
func GetTimeProvider(name string) TimeProvider {
	provider, ok := timeProviders[name]
	if ok {
		return provider
	}
	return nil
}

// CleanTimeProviderConfig ...
func CleanTimeProviderConfig(data interface{}, name string) interface{} {
	var connection Connection
	connection = GetTimeProvider(name).New(data)
	return connection.CleanTimeProviderConfig()
}

// GetProjectTimeLogs ...
func GetProjectTimeLogs(
	config []byte,
	project string,
	startTime, endTime time.Time) ([]serializers.TimeLog, error) {
	connections, err := GetConnections(config)
	if err != nil {
		return nil, err
	}
	timeLogs := make([]serializers.TimeLog, 0)

	for _, connection := range connections {
		timeLogs = append(timeLogs, connection.GetProjectTimeLogs(project, startTime, endTime)...)
	}
	return timeLogs, nil
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
		connection = GetTimeProvider(name).New(data)
		if connection != nil {
			connections = append(connections, connection)
		}
	}
	return connections, nil
}
