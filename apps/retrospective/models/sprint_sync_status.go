package models

import (
	"errors"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/sirupsen/logrus"
)

// SyncStatusValues ...
var SyncStatusValues = [...]string{
	"Not Synced",
	"Syncing",
	"Synced",
	"Sync Failed",
	"Queued",
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
	Queued
)

// SprintSyncStatus stores the sync history of a sprint
type SprintSyncStatus struct {
	gorm.Model
	SprintID uint `gorm:"not null"`
	Sprint   Sprint
	Status   SyncStatus `gorm:"default:0; not null"`
}

// Validate ...
func (syncStatus *SprintSyncStatus) Validate(db *gorm.DB) (err error) {
	if syncStatus.Status < 0 || int(syncStatus.Status) > len(SyncStatusValues) {
		err = errors.New("please select a valid sync status")
		return err
	}
	return
}

// BeforeSave ...
func (syncStatus *SprintSyncStatus) BeforeSave(db *gorm.DB) (err error) {
	return syncStatus.Validate(db)
}

// BeforeUpdate ...
func (syncStatus *SprintSyncStatus) BeforeUpdate(db *gorm.DB) (err error) {
	return syncStatus.Validate(db)
}

// RegisterSprintSyncStatusToAdmin ...
func RegisterSprintSyncStatusToAdmin(Admin *admin.Admin, config admin.Config) {
	userTeam := Admin.AddResource(&SprintSyncStatus{}, &config)
	roleMeta := getSprintSyncStatusMeta()
	userTeam.Meta(&roleMeta)
}

func getSprintSyncStatusMeta() admin.Meta {
	return admin.Meta{
		Name: "Status",
		Type: "select_one",
		Valuer: func(value interface{}, context *qor.Context) interface{} {
			syncStatus := value.(*SprintSyncStatus)
			return strconv.Itoa(int(syncStatus.Status))
		},
		Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			syncStatus := resource.(*SprintSyncStatus)
			value, err := strconv.Atoi(metaValue.Value.([]string)[0])
			if err != nil {
				logrus.Error("Cannot convert string to int")
				return
			}
			syncStatus.Status = SyncStatus(value)
		},
		Collection: func(value interface{}, context *qor.Context) (results [][]string) {
			for index, value := range SyncStatusValues {
				results = append(results, []string{strconv.Itoa(index), value})
			}
			return
		},
		FormattedValuer: func(value interface{}, context *qor.Context) interface{} {
			syncStatus := value.(*SprintSyncStatus)
			return syncStatus.Status.GetStringValue()
		},
	}
}
