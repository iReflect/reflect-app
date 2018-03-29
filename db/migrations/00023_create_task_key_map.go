package migrations

import (
	"database/sql"
	"github.com/iReflect/reflect-app/db/base/models"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00023, Down00023)
}

// Up00023 ...
func Up00023(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&models.TaskKeyMap{})
	gormdb.Model(&models.TaskKeyMap{}).AddForeignKey("task_id", "tasks(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.TaskKeyMap{}).AddUniqueIndex("unique_task_id_key", "task_id", "key")
	gormdb.Model(&models.TaskKeyMap{}).AddIndex("idx_task_key", "key")

	return nil
}

// Down00023 ...
func Down00023(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.TaskKeyMap{}).RemoveIndex("idx_task_key")
	gormdb.Model(&models.TaskKeyMap{}).RemoveIndex("unique_task_id_key")
	gormdb.Model(&models.TaskKeyMap{}).RemoveForeignKey("task_id", "tasks(id)")
	gormdb.DropTable(&models.TaskKeyMap{})

	return nil
}
