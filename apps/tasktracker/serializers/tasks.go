package serializers

import (
	"time"
)

//Task ...
type Task struct {
	ID        string
	ProjectID string
	Summary   string
	Type      string
	Priority  string
	Estimate  *float64
	Assignee  string
	Status    string
}

//Sprint ...
type Sprint struct {
	ID       string
	BoardID  string
	Name     string
	FromDate time.Time
	ToDate   time.Time
}

//Board ...
type Board struct {
	ID   string
	Name string
}
