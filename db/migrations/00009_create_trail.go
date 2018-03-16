package migrations

import (
	"database/sql"
	"github.com/iReflect/reflect-app/db/base/models"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00009, Down00009)
}

// Up00009 ...
func Up00009(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&models.Trail{})

	gormdb.Model(&models.Trail{}).AddForeignKey("action_by_id", "users(id)", "RESTRICT", "RESTRICT")

	gormdb.Model(&models.Trail{}).AddIndex("index_trails_action_item_action_item_id", "action_item", "action_item_id")

	return nil
}

// Down00009 ...
func Down00009(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.Trail{}).RemoveIndex("index_trails_action_item_action_item_id")

	gormdb.Model(&models.Trail{}).RemoveForeignKey("action_by_id", "users(id)")

	gormdb.DropTable(&models.Trail{})

	return nil
}
