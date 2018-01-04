package serializers

import (
	"time"
)

type TimeLog struct {
	ID        string
	ProjectID string
	TaskID    string
	Summary   string
	Logger    string
	Minutes   int
	Date      time.Time
}

type Project struct {
	ID   string
	Name *string
}
