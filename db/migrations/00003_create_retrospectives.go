package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00003, Down00003)
}

// Up00003 ...
func Up00003(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&models.Retrospective{})

	gormdb.Model(&models.Retrospective{}).AddForeignKey("team_id", "teams(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.Retrospective{}).AddForeignKey("created_by_id", "users(id)", "RESTRICT", "RESTRICT")

	gormdb.Model(&models.Retrospective{}).AddIndex("idx_retrospectives_team_id", "team_id")

	return nil
}

// Down00003 ...
func Down00003(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.Retrospective{}).RemoveIndex("idx_retrospectives_team_id")

	gormdb.Model(&models.Retrospective{}).RemoveForeignKey("team_id", "teams(id)")
	gormdb.Model(&models.Retrospective{}).RemoveForeignKey("created_by_id", "users(id)")

	gormdb.DropTable(&models.Retrospective{})

	return nil
}
