package migrations

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/apps/retrospective/models"
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

	memberTaskRoleValuesLength := len(models.MemberTaskRoleValues)
	gormdb.Exec("UPDATE sprint_member_tasks SET role = ? where role >= ?", memberTaskRoleValuesLength-1, memberTaskRoleValuesLength)

	return nil
}

// Down00028 ...
func Down00028(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.

	return nil
}
