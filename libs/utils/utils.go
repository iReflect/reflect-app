package utils

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"time"
)

// RandToken ...
func RandToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// LogToSentry ...
func LogToSentry(err error) {
	logrus.Error(err.Error())
	// ToDo: Add extra info like release, etc
	raven.CaptureError(err, nil)
}

// UIntInSlice ...
func UIntInSlice(element uint, slice []uint) bool {
	for _, sliceElement := range slice {
		if sliceElement == element {
			return true
		}
	}
	return false
}

// EncryptionKey ...
func EncryptionKey() []byte {
	key := os.Getenv("ENCRYPTION_KEY")
	if len(key) == 0 {
		key = "DUMMY_KEY__FOR_LOCAL_DEV"
	}
	return []byte(key)
}

// GetWorkingDaysBetweenTwoDates calculates the working days between two dates,
// i.e., number of days between two dates excluding weekends
func GetWorkingDaysBetweenTwoDates(startDate time.Time, endDate time.Time) int {
	if endDate.Before(startDate) {
		return 0
	}
	workingDays := 0
	start := startDate
	end := endDate
	for start.Weekday() != time.Monday && start.Before(end) {
		if start.Weekday() != time.Sunday && start.Weekday() != time.Saturday {
			workingDays++
		}
		start = start.Add(time.Hour * 24)
	}

	for end.Weekday() != time.Sunday && end.After(start) {
		if end.Weekday() != time.Saturday {
			workingDays++
		}
		end = end.Add(-time.Hour * 24)
	}
	duration := end.Sub(start)
	if duration.Hours() > 24 {
		weeks := int(math.Ceil(duration.Hours() / (24*7)))
		workingDays += weeks*5
	}  else {
		workingDays++
	}

	return workingDays
}

// CalculateExpectedSP ...
func CalculateExpectedSP(startDate time.Time, endDate time.Time, vacations float64, expectationPercent float64, allocationPercent float64, spPerWeek float64) float64 {
	sprintWorkingDays := GetWorkingDaysBetweenTwoDates(startDate, endDate)
	workingDays := float64(sprintWorkingDays) - vacations
	expectationCoefficient := expectationPercent / 100.00
	allocationCoefficient := allocationPercent / 100.00
	storyPointPerDay := spPerWeek / 5
	return workingDays * storyPointPerDay * expectationCoefficient * allocationCoefficient
}

// StringSliceToInterfaceSlice ...
func StringSliceToInterfaceSlice(originalSlice []string) []interface{} {
	newSlice := make([]interface{}, len(originalSlice))
	for i, v := range originalSlice {
		newSlice[i] = v
	}
	return newSlice
}

// InterfaceSliceToStringSlice ...
func InterfaceSliceToStringSlice(originalSlice []interface{}) []string {
	newSlice := make([]string, len(originalSlice))
	for i, v := range originalSlice {
		newSlice[i] = v.(string)
	}
	return newSlice
}
