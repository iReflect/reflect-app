package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00019, Down00019)
}

// Up00019 ...
func Up00019(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&models.SprintSyncStatus{})

	gormdb.Model(&models.SprintSyncStatus{}).AddForeignKey("sprint_id", "sprints(id)", "RESTRICT", "RESTRICT")

	return nil
}

// Down00019 ...
func Down00019(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.SprintSyncStatus{}).RemoveForeignKey("sprint_id", "sprints(id)")

	gormdb.DropTable(&models.SprintSyncStatus{})

	return nil
}
