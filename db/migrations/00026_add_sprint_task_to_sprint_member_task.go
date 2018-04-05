package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00026, Down00026)
}

// Up00026 ...
func Up00026(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	type sprintMemberTask struct {
		gorm.Model
		SprintTaskID uint `gorm:"not null"`
	}

	gormdb.AutoMigrate(&sprintMemberTask{})
	gormdb.Model(&sprintMemberTask{}).AddForeignKey("sprint_task_id", "sprint_tasks(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&sprintMemberTask{}).RemoveForeignKey("task_id", "tasks(id)")
	gormdb.Model(&sprintMemberTask{}).DropColumn("task_id")

	return nil
}

// Down00026 ...
func Down00026(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type sprintMemberTask struct {
		gorm.Model
		TaskID uint `gorm:"not null"`
	}

	gormdb.AutoMigrate(&sprintMemberTask{})
	gormdb.Model(&sprintMemberTask{}).AddForeignKey("task_id", "tasks(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&sprintMemberTask{}).RemoveForeignKey("sprint_task_id", "sprint_tasks(id)")
	gormdb.Model(&sprintMemberTask{}).DropColumn("sprint_task_id")

	return nil
}
