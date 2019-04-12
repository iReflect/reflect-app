package timetracker

import (
	"strings"
	"time"

	"encoding/json"

	"github.com/iReflect/reflect-app/apps/timetracker/serializers"
)

// TimeProvider ...
type TimeProvider interface {
	New(config interface{}) Connection
}

// Connection ...
type Connection interface {
	GetProjectTimeLogs(project string, startTime time.Time, endTime time.Time) []serializers.TimeLog
}

var timeProviders = make(map[string]TimeProvider)

// RegisterTimeProvider ...
func RegisterTimeProvider(name string, newProvider TimeProvider) {
	timeProviders[name] = newProvider
}

// GetTimeProvider ...
func GetTimeProvider(name string) TimeProvider {
	provider, ok := timeProviders[name]
	if ok {
		return provider
	}
	return nil
}

// GetProjectTimeLogs ...
func GetProjectTimeLogs(config []byte, project string, startTime time.Time, endTime time.Time) (timeLogs []serializers.TimeLog, err error) {
	connections, err := GetConnections(config)
	if err != nil {
		return nil, err
	}

	for _, connection := range connections {
		timeLogs = append(timeLogs, connection.GetProjectTimeLogs(project, startTime, endTime)...)
	}
	CleanTickets(timeLogs)
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

// CleanTickets ...
func CleanTickets(timeLogs []serializers.TimeLog) {
	for index, value := range timeLogs {
		timeLogs[index].TaskKey = strings.TrimPrefix(value.TaskKey, "#")
	}
}
