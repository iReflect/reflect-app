package migrations

import (
	"database/sql"
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00029, Down00029)
}

// Up00029 ...
func Up00029(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type task struct {
		Rating int8 `gorm:"default:2; not null"`
	}

	gormdb.AutoMigrate(&task{})

	return nil
}

// Down00029 ...
func Down00029(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.Task{}).DropColumn("rating")

	return nil
}
