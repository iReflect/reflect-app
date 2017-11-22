package migrations

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

// Define only the fields used in this migration and not full model.
type Category struct {
	gorm.Model
	Weight int
}
type ItemType struct {
	gorm.Model
	Weight int
}

func init() {
	goose.AddMigration(Up20171103004441, Down20171103004441)
}

func Up20171103004441(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&Category{})
	gormdb.Model(&ItemType{}).DropColumn("weight")

	return nil
}

func Down20171103004441(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&ItemType{})
	gormdb.Model(&Category{}).DropColumn("weight")

	return nil
}
