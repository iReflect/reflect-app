package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00010, Down00010)
}

// Up00010 ...
func Up00010(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	//sprint (only the fields used in this migration and not full model)
	type sprint struct {
		SprintLearnings  string `gorm:"type:text"`
		SprintHighlights string `gorm:"type:text"`
	}

	// Automigrate sprint Model
	gormdb.AutoMigrate(&sprint{})

	return nil
}

// Down00010 ...
func Down00010(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	//sprint (only the fields used in this migration and not full model)
	type sprint struct {
		SprintLearnings  string `gorm:"type:text"`
		SprintHighlights string `gorm:"type:text"`
	}

	// Drop added columns
	gormdb.Model(&sprint{}).DropColumn("sprint_learnings")
	gormdb.Model(&sprint{}).DropColumn("sprint_highlights")

	return nil
}
