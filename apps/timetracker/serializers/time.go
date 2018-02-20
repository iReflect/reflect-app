package serializers

import (
	"time"
)

//TimeLog ...
type TimeLog struct {
	ID        string
	ProjectID string
	TaskID    string
	Summary   string
	Logger    string
	Minutes   uint
	Date      time.Time
}

//Project ...
type Project struct {
	ID   string
	Name *string
}
