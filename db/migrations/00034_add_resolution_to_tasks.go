package migrations

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00034, Down00034)
}

// Up00034 ...
func Up00034(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormDB, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type task struct {
		Resolution int8 `gorm:"default:0"`
	}

	err = gormDB.AutoMigrate(&task{}).Error
	if err != nil {
		return err
	}
	err = gormDB.Model(retroModels.Task{}).
		Where("tasks.deleted_at IS NULL").
		Not("tasks.done_at IS NULL").
		Update("resolution", retroModels.DoneResolution).
		Error
	if err != nil {
		return err
	}

	return nil
}

// Down00034 ...
func Down00034(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormDB, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	err = gormDB.Model(&models.Task{}).DropColumn("resolution").Error
	if err != nil {
		return err
	}
	return nil
}
