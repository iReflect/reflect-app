package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00022, Down00022)
}

type task struct {
	TrackerUniqueID string `gorm:"type:varchar(255); not null; default:''"`
	Description     string `gorm:"type:text; not null; default:''"`
}

// Up00022 ...
func Up00022(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&task{})

	return nil
}

// Down00022 ...
func Down00022(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	// Drop a column
	gormdb.Model(&task{}).DropColumn("tracker_unique_id")
	gormdb.Model(&task{}).DropColumn("description")

	return nil
}
