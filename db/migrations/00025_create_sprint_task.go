package migrations

import (
	"database/sql"
	"github.com/iReflect/reflect-app/db/base/models"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00025, Down00025)
}

// Up00025 ...
func Up00025(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&models.SprintTask{})
	gormdb.Model(&models.SprintTask{}).AddForeignKey("task_id", "tasks(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.SprintTask{}).AddForeignKey("sprint_id", "sprints(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.SprintTask{}).AddIndex("idx_sprint_tasks__sprint_id__task_id", "sprint_id", "task_id")

	return nil
}

// Down00025 ...
func Down00025(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.SprintTask{}).RemoveIndex("idx_sprint_tasks__sprint_id__task_id")
	gormdb.Model(&models.SprintTask{}).RemoveForeignKey("sprint_id", "sprints(id)")
	gormdb.Model(&models.SprintTask{}).RemoveForeignKey("task_id", "tasks(id)")
	gormdb.DropTable(&models.SprintTask{})

	return nil
}
