package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00012, Down00012)
}

// Up00012 ...
func Up00012(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type sprint struct {
		SprintLearnings  string `gorm:"type:text"`
		SprintHighlights string `gorm:"type:text"`
	}

	// Drop added columns
	gormdb.Model(&sprint{}).DropColumn("sprint_learnings")
	gormdb.Model(&sprint{}).DropColumn("sprint_highlights")

	return nil
}

// Down00012 ...
func Down00012(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type sprint struct {
		SprintLearnings  string `gorm:"type:text"`
		SprintHighlights string `gorm:"type:text"`
	}

	// Automigrate sprint Model
	gormdb.AutoMigrate(&sprint{})

	return nil
}
