package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00005, Down00005)
}

// Up00005 ...
func Up00005(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&models.Task{})

	gormdb.Model(&models.Task{}).AddForeignKey("retrospective_id", "retrospectives(id)", "RESTRICT", "RESTRICT")

	gormdb.Model(&models.Task{}).AddUniqueIndex("unique_tasks_retro_id_task_id", "retrospective_id", "task_id")

	return nil
}

// Down00005 ...
func Down00005(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.Task{}).RemoveIndex("unique_tasks_retro_id_task_id")

	gormdb.Model(&models.Task{}).RemoveForeignKey("retrospective_id", "retrospectives(id)")

	gormdb.DropTable(&models.Task{})

	return nil
}
