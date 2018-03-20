package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00020, Down00020)
}

// Up00020 ...
func Up00020(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Exec("ALTER TABLE tasks ALTER COLUMN estimate SET DEFAULT 0;")
	gormdb.Exec("UPDATE tasks SET estimate = 0 WHERE estimate IS NULL;")
	gormdb.Exec("ALTER TABLE tasks ALTER COLUMN estimate SET NOT NULL;")

	return nil
}

// Down00020 ...
func Down00020(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Exec("ALTER TABLE tasks ALTER COLUMN estimate DROP NOT NULL;")
	gormdb.Exec("ALTER TABLE tasks ALTER COLUMN estimate DROP DEFAULT;")

	return nil
}
