package providers

import (
	"time"

	"github.com/iReflect/reflect-app/apps/timetracker"
	"github.com/iReflect/reflect-app/apps/timetracker/serializers"
)

//GSheetTimeProvider ...
type GSheetTimeProvider struct {
}

//GsheetConnection ...
type GsheetConnection struct {
	config GsheetConfig
}

//GsheetConfig ...
type GsheetConfig struct {
	Credentials string
	SheetID     string
}

// TimeProviderGSheet ...
const (
	TimeProviderGSheet = "gsheet"
)

func init() {
	provider := &GSheetTimeProvider{}
	timetracker.RegisterTimeProvider(TimeProviderGSheet, provider)
}

//New ...
func (m *GSheetTimeProvider) New(config interface{}) timetracker.Connection {
	c, ok := config.(GsheetConfig)
	if !ok {
		return nil
	}
	return &GsheetConnection{config: c}
}

//GetProjects ...
func (m *GsheetConnection) GetProjects() []serializers.Project {
	return nil
}

//GetTimeLogs ...
func (m *GsheetConnection) GetTimeLogs(startTime time.Time, endTime time.Time) []serializers.TimeLog {
	return nil
}

//GetProjectTimeLogs ...
func (m *GsheetConnection) GetProjectTimeLogs(project serializers.Project, startTime time.Time, endTime time.Time) []serializers.TimeLog {
	return nil
}
