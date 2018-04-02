package migrations

import (
	"database/sql"
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
	"time"
)

func init() {
	goose.AddMigration(Up00021, Down00021)
}

// Up00021 ...
func Up00021(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type task struct {
		DoneAt *time.Time
	}

	gormdb.AutoMigrate(&task{})

	gormdb.Model(&models.Task{}).AddIndex("idx_tasks_done_at", "done_at")

	return nil
}

// Down00021 ...
func Down00021(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.Task{}).RemoveIndex("idx_tasks_done_at")

	gormdb.Model(&models.Task{}).DropColumn("done_at")

	return nil
}
