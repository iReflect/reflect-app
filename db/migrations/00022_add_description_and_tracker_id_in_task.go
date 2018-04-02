package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00022, Down00022)
}

// Up00022 ...
func Up00022(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type task struct {
		TrackerUniqueID string `gorm:"type:varchar(255); not null; default:''"`
		Description     string `gorm:"type:text; not null; default:''"`
	}

	gormdb.AutoMigrate(&task{})
	gormdb.Model(&task{}).AddIndex("index_tasks_retro_id_tu_id", "retrospective_id", "tracker_unique_id")
	return nil
}

// Down00022 ...
func Down00022(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type task struct {
	}

	// Drop a column
	gormdb.Model(&task{}).RemoveIndex("index_tasks_retro_id_tu_id")
	gormdb.Model(&task{}).DropColumn("tracker_unique_id")
	gormdb.Model(&task{}).DropColumn("description")

	return nil
}
