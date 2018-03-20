package migrations

import (
	"database/sql"
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00015, Down00015)
}

// Up00015 ...
func Up00015(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type sprintMemberTask struct {
		gorm.Model
		Role int8 `gorm:"default:0; not null"`
	}

	gormdb.AutoMigrate(&sprintMemberTask{})

	return nil
}

// Down00015 ...
func Down00015(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.SprintMemberTask{}).DropColumn("role")

	return nil
}
