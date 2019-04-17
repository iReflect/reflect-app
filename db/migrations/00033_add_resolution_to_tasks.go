package migrations

import (
	"database/sql"

	retroModels "github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/db/base/models"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00033, Down00033)
}

// Up00033 ...
func Up00033(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type task struct {
		Resolution int8 `gorm:"default:1"`
	}
	gormdb.AutoMigrate(&task{})

	err = gormdb.Model(retroModels.Task{}).
		Where("tasks.deleted_at IS NULL").
		Where("tasks.done_at IS NULL").
		Update("resolution", nil).
		Error
	if err != nil {
		return err
	}

	return nil
}

// Down00033 ...
func Down00033(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.Task{}).DropColumn("resolution")

	return nil
}
