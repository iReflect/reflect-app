package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/models/fields"
)

// User ...
type User struct {
	TimeProviderConfig fields.JSONB `gorm:"type:jsonb; not null; default:'{}'::jsonb"`
}

func init() {
	goose.AddMigration(Up00008, Down00008)
}

// Up00008 ...
func Up00008(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&User{})

	return nil
}

// Down00008 ...
func Down00008(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&User{}).DropColumn("time_provider_config")

	return nil
}
