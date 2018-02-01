package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00007, Down00007)
}

// Up00007 ...
func Up00007(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&models.SprintMemberTask{})

	gormdb.Model(&models.SprintMemberTask{}).AddForeignKey("sprint_member_id", "sprint_members(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.SprintMemberTask{}).AddForeignKey("task_id", "tasks(id)", "RESTRICT", "RESTRICT")

	gormdb.Model(&models.SprintMemberTask{}).AddUniqueIndex("unique_sprint_member_tasks_sprint_member_id_task_id", "sprint_member_id", "task_id")

	return nil
}

// Down00007 ...
func Down00007(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.SprintMemberTask{}).RemoveIndex("unique_sprint_member_tasks_sprint_member_id_task_id")

	gormdb.Model(&models.SprintMemberTask{}).RemoveForeignKey("sprint_member_id", "sprint_members(id)")
	gormdb.Model(&models.SprintMemberTask{}).RemoveForeignKey("task_id", "tasks(id)")

	gormdb.DropTable(&models.SprintMemberTask{})

	return nil
}
