package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00018, Down00018)
}

// Up00018 ...
func Up00018(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	// Drop added columns
	gormdb.Model(&models.Sprint{}).DropColumn("currently_syncing")

	return nil
}

// Down00018 ...
func Down00018(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type sprint struct {
		CurrentlySyncing bool `gorm:"default:true;not null"`
	}
	// Automigrate sprint Model
	gormdb.AutoMigrate(&sprint{})

	return nil
}
