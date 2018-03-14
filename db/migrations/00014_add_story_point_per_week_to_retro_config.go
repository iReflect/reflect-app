package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00014, Down00014)
}

// Up00014 ...
func Up00014(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	type retrospective struct {
		gorm.Model
		StoryPointPerWeek float64 `gorm:"not null"`
	}

	gormdb.AutoMigrate(&retrospective{})
	gormdb.Exec("UPDATE retrospectives SET story_point_per_week = (40/COALESCE(hrs_per_story_point,4))")
	gormdb.Model(&retrospective{}).DropColumn("hrs_per_story_point")
	return nil
}

// Down00014 ...
func Down00014(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	type retrospective struct {
		gorm.Model
		HrsPerStoryPoint float64 `gorm:"not null"`
	}

	gormdb.AutoMigrate(&retrospective{})
	gormdb.Model(&retrospective{}).DropColumn("story_point_per_week")

	return nil
}
