package migrations

import (
	"database/sql"

	"github.com/iReflect/reflect-app/db/base/models"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00033, Down00033)
}

// Up00033 ...
func Up00033(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type Team struct {
		TimeProviderName string `gorm:"default:'gsheet'; not null"`
	}

	type Retrospective struct {
		TimeProviderName string `gorm:"default:'gsheet'; not null"`
	}

	gormdb.AutoMigrate(&Team{}, &Retrospective{})

	return nil
}

// Down00033 ...
func Down00033(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.Team{}).DropColumn("time_provider_name")
	gormdb.Model(&models.Retrospective{}).DropColumn("time_provider_name")

	return nil
}
