package models

import (
	"github.com/jinzhu/gorm"
)

// SyncStatusValues ...
var SyncStatusValues = [...]string{
	"Not Synced",
	"Syncing",
	"Synced",
	"Sync Failed",
}

// SyncStatus ...
type SyncStatus int8

// GetStringValue ...
func (status SyncStatus) GetStringValue() string {
	return SyncStatusValues[status]
}

// SyncStatus
const (
	NotSynced SyncStatus = iota
	Syncing
	Synced
	SyncFailed
)

// SprintSyncStatus stores the sync history of a sprint
type SprintSyncStatus struct {
	gorm.Model
	SprintID uint `gorm:"not null"`
	Sprint   Sprint
	Status   SyncStatus `gorm:"default:0; not null"`
}
