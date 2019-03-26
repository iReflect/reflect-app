package migrations

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00030, Down00030)
}

// Up00030 ...
func Up00030(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormDB, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	gormDB.Model(&models.Task{}).ModifyColumn("summary", "text")

	return nil
}

// Down00030 ...
func Down00030(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormDB, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	gormDB.Model(&models.Task{}).ModifyColumn("summary", "varchar(255)")

	return nil
}
