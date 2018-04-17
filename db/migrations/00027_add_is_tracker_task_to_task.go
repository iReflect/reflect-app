package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00027, Down00027)
}

// Up00027 ...
func Up00027(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	type task struct {
		IsTrackerTask bool `gorm:"not null; default: false"`
	}

	gormdb.AutoMigrate(&task{})

	return nil
}

// Down00027 ...
func Down00027(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.Task{}).DropColumn("is_tracker_task")

	return nil
}
