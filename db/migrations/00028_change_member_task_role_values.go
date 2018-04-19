package migrations

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00028, Down00028)
}

// Up00028 ...
func Up00028(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	// Changed Member Task Role from Validator to Reviewer
	gormdb.Table("sprint_member_tasks").Where("role = 2").Update("role", 1)

	return nil
}

// Down00028 ...
func Down00028(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.

	return nil
}
