package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00004, Down00004)
}

// Up00004 ...
func Up00004(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&models.Sprint{})

	gormdb.Model(&models.Sprint{}).AddForeignKey("retrospective_id", "retrospectives(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.Sprint{}).AddForeignKey("created_by_id", "users(id)", "RESTRICT", "RESTRICT")

	gormdb.Model(&models.Sprint{}).AddIndex("idx_sprints_retro_id_end_date", "retrospective_id", "end_date")
	gormdb.Model(&models.Sprint{}).AddIndex("idx_sprints_retro_id_status_end_date", "retrospective_id", "status", "end_date")

	return nil
}

// Down00004 ...
func Down00004(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.Sprint{}).RemoveIndex("idx_sprints_retro_id_end_date")
	gormdb.Model(&models.Sprint{}).RemoveIndex("idx_sprints_retro_id_status_end_date")

	gormdb.Model(&models.Sprint{}).RemoveForeignKey("created_by_id", "users(id)")
	gormdb.Model(&models.Sprint{}).RemoveForeignKey("retrospective_id", "retrospectives(id)")

	gormdb.DropTable(&models.Sprint{})

	return nil
}
