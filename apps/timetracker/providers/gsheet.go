package providers

import (
	"time"

	"github.com/iReflect/reflect-app/apps/timetracker"
	"github.com/iReflect/reflect-app/apps/timetracker/serializers"
)

//GSheetTimeProvider ...
type GSheetTimeProvider struct {
}

type GsheetConnection struct {
	config GsheetConfig
}

type GsheetConfig struct {
	Credentials string
	SheetID     string
}

const (
	TIME_PROVIDER_GSHEET = "gsheet"
)

func init() {
	provider := &GSheetTimeProvider{}
	timetracker.RegisterTimeProvider(TIME_PROVIDER_GSHEET, provider)
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
