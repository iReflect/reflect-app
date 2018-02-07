package timetracker

import (
	"time"

	"github.com/iReflect/reflect-app/apps/timetracker/serializers"
)

// TimeProvider ...
type TimeProvider interface {
	New(config interface{}) Connection
}

// Connection ...
type Connection interface {
	GetProjects() []serializers.Project
	GetTimeLogs(startTime time.Time, endTime time.Time) []serializers.TimeLog
	GetProjectTimeLogs(project serializers.Project, startTime time.Time, endTime time.Time) []serializers.TimeLog
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
