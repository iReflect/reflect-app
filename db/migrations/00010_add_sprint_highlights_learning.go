package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

//Sprint (only the fields used in this migration and not full model)
type Sprint struct {
	SprintLearnings  string `gorm:"type:text"`
	SprintHighlights string `gorm:"type:text"`
}

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

	// Automigrate Sprint Model
	gormdb.AutoMigrate(&Sprint{})

	return nil
}

// Down00010 ...
func Down00010(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	// Drop added columns
	gormdb.Model(&Sprint{}).DropColumn("sprint_learnings")
	gormdb.Model(&Sprint{}).DropColumn("sprint_highlights")

	return nil
}
