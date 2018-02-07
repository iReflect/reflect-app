package serializers

import (
	"encoding/json"
	"time"
)

//Task ...
type Task struct {
	ID          string
	ProjectID   string
	Summary     string
	Description string
	Type        string
	Priority    string
	Estimate    *int
	Assignee    string
	Status      string
	SprintID    *string
	Fields      *json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
