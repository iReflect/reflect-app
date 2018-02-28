package providers

import (
	"time"

	"encoding/json"
	"errors"
	"github.com/iReflect/reflect-app/apps/timetracker"
	"github.com/iReflect/reflect-app/apps/timetracker/serializers"
	"github.com/iReflect/reflect-app/libs/sheets"
)

// GSheetTimeProvider ...
type GSheetTimeProvider struct {
}

// GsheetConnection ...
type GsheetConnection struct {
	config GsheetConfig
}

// GsheetConfig ...
type GsheetConfig struct {
	URL     string `json:"URL"`
	SheetID string `json:"SheetID"`
}

// TimeProviderGSheet ...
const (
	TimeProviderGSheet = "gsheet"
)

func init() {
	provider := &GSheetTimeProvider{}
	timetracker.RegisterTimeProvider(TimeProviderGSheet, provider)
}

// New ...
func (m *GSheetTimeProvider) New(config interface{}) timetracker.Connection {
	var gsheetConfig GsheetConfig
	gsheetConfig, err := getConfigObject(config)

	if err != nil {
		return nil
	}

	return &GsheetConnection{config: gsheetConfig}
}

// getConfigObject ...
func getConfigObject(config interface{}) (GsheetConfig, error) {
	var c GsheetConfig

	switch config.(type) {
	case []byte:
		c = GsheetConfig{}
		err := json.Unmarshal(config.([]byte), &c)
		if err != nil {
			return c, err
		}
	case map[string]interface{}:
		c = GsheetConfig{}

		jsonConfig, err := json.Marshal(config)
		if err != nil {
			return c, err
		}

		err = json.Unmarshal(jsonConfig, &c)
		if err != nil {
			return c, err
		}
	case GsheetConfig:
		c = config.(GsheetConfig)
	default:
		return c, errors.New("Invalid type")
	}
	return c, nil
}

// GetProjectTimeLogs ...
func (m *GsheetConnection) GetProjectTimeLogs(project string, startTime time.Time, endTime time.Time) []serializers.TimeLog {
	trackerData := sheets.GetTimeData(project, m.config.SheetID, startTime.String(), endTime.String())
	var timeLogs []serializers.TimeLog

	for _, logData := range trackerData {
		timeLogs = append(timeLogs, serializers.TimeLog{
			Project: logData["Project"].(string),
			TaskID:  logData["TaskID"].(string),
			Logger:  "GSheets",
			Minutes: uint(logData["Hours"].(float64) * 60),
		})
	}

	return timeLogs
}
