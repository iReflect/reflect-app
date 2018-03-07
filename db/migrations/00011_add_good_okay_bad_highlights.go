package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00011, Down00011)
}

// Up00011 ...
func Up00011(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type sprint struct {
		GoodHighlights string `gorm:"type:text"`
		OkayHighlights string  `gorm:"type:text"`
		BadHighlights  string `gorm:"type:text"`
	}

	// Automigrate sprint Model
	gormdb.AutoMigrate(&sprint{})

	return nil
}

// Down00011 ...
func Down00011(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type sprint struct {
		GoodHighlights string `gorm:"type:text"`
		OkayHighlights string  `gorm:"type:text"`
		BadHighlights  string `gorm:"type:text"`
	}

	// Drop added columns
	gormdb.Model(&sprint{}).DropColumn("good_highlights")
	gormdb.Model(&sprint{}).DropColumn("bad_highlights")
	gormdb.Model(&sprint{}).DropColumn("ugly_highlights")

	return nil
}
