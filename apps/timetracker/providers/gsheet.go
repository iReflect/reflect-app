package providers

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/iReflect/reflect-app/apps/timetracker"
	"github.com/iReflect/reflect-app/apps/timetracker/serializers"
	"github.com/iReflect/reflect-app/config"
	"github.com/iReflect/reflect-app/constants"
	"github.com/iReflect/reflect-app/libs/google"
	"github.com/iReflect/reflect-app/libs/utils"
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
	Email string `json:"email"`
}

// TimeResult ...
type TimeResult struct {
	Project string  `json:"Project"`
	TaskID  string  `json:"TaskID"`
	Hours   float64 `json:"Hours"`
}

// TimeProviderGSheet ...
const (
	TimeProviderGSheet            = "gsheet"
	TimeProviderGSheetDisplayName = "Google Timesheet"
)

func init() {
	provider := &GSheetTimeProvider{}
	timetracker.RegisterTimeProvider(TimeProviderGSheet, provider)
	timetracker.RegisterTimeProviderDisplayName(TimeProviderGSheet, TimeProviderGSheetDisplayName)
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
		return c, errors.New("invalid type")
	}
	return c, nil
}

// CleanTimeProviderConfig ...
func (m *GsheetConnection) CleanTimeProviderConfig() interface{} {
	m.config.Email = strings.TrimSpace(m.config.Email)
	return m.config
}

// GetProjectTimeLogs ...
func (m *GsheetConnection) GetProjectTimeLogs(project string, startTime time.Time, endTime time.Time) []serializers.TimeLog {

	timeLogs := make([]serializers.TimeLog, 0)
	timeTrackerConfig := config.GetConfig().TimeTracker
	appExecutor := google.AppScriptExecutor{ScriptID: timeTrackerConfig.ScriptID, CredentialsFile: timeTrackerConfig.GoogleCredentials}

	location, err := time.LoadLocation(timeTrackerConfig.TimeZone)
	if err != nil {
		log.Println("Invalid Timezone: ", err)
		utils.LogToSentry(err)
		return timeLogs
	}
	responseBytes, err := appExecutor.Run(
		timeTrackerConfig.FnGetTimeLog,
		m.config.Email,
		project, startTime.In(location).Format(constants.CustomDateFormat),
		endTime.In(location).Format(constants.CustomDateFormat))
	if err != nil {
		log.Println("App Executor Failed: ", err)
		utils.LogToSentry(err)
		return timeLogs
	}

	type Response struct {
		Type   string       `json:"type"`
		Result []TimeResult `json:"result"`
	}

	var trackerData Response

	if err := json.Unmarshal(responseBytes, &trackerData); err != nil {
		log.Println("Respoonse decoding error: ", err)
		utils.LogToSentry(err)
		return timeLogs
	}

	log.Println("Result : ", trackerData.Result)

	for _, logData := range trackerData.Result {
		if logData.TaskID != "" {
			timeLogs = append(timeLogs, serializers.TimeLog{
				Project: logData.Project,
				TaskKey: logData.TaskID,
				Logger:  "GSheets",
				Minutes: uint(logData.Hours * 60), //uint(logData["Hours"].(float64) * 60),
				Email:   m.config.Email,
			})
		}
	}

	return timeLogs
}
