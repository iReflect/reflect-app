package utils

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
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
func GetWorkingDaysBetweenTwoDates(startDate time.Time, endDate time.Time, includeBoth bool) int {
	if endDate.Before(startDate) {
		return -1
	}
	workingDays := 0
	startDay := startDate.Weekday()
	endDay := endDate.Weekday()

	// normalize dates to calculate time difference
	startDate = startDate.AddDate(0, 0, int(-startDay))
	endDate = endDate.AddDate(0, 0, int(-endDay))

	diffDays := endDate.Sub(startDate).Hours() / 24
	daysWithoutWeekendDays := int(diffDays - (diffDays * 2 / 7))

	if includeBoth && ((startDay != time.Saturday && startDay != time.Sunday) || (endDay != time.Saturday && endDay != time.Sunday)) {
		workingDays++
	}

	// normalize start day to account for saturday/sunday
	if startDay == time.Sunday && endDay != time.Saturday {
		startDay = time.Monday
	} else if startDay == time.Saturday && endDay != time.Sunday {
		startDay = time.Friday
	}

	// normalize end day to account for saturday/sunday
	if endDay == time.Sunday {
		endDay = time.Monday
	} else if endDay == time.Saturday {
		endDay = time.Friday
	}

	workingDays += daysWithoutWeekendDays - int(startDay) + int(endDay)

	return workingDays
}
