package migrations

import (
	"database/sql"
	"github.com/iReflect/reflect-app/db/base/models"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00024, Down00024)
}

// Up00024 ...
func Up00024(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.Task{}).RemoveIndex("index_tasks_retro_id_task_id")
	gormdb.Exec("ALTER TABLE tasks RENAME COLUMN task_id to key")

	return nil
}

// Down00024 ...
func Down00024(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Exec("ALTER TABLE tasks RENAME COLUMN key to task_id")
	gormdb.Model(&models.Task{}).AddIndex("index_tasks_retro_id_task_id", "retrospective_id", "task_id")

	return nil
}
