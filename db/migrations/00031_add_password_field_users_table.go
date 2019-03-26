package migrations

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00031, Down00031)
}

// Up00031 ...
func Up00031(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormDB, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type User struct {
		Password []byte `gorm:"type:bytea"`
	}

	gormDB.AutoMigrate(&User{})

	return nil
}

// Down00031 ...
func Down00031(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormDB, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormDB.Model(&models.User{}).DropColumn("password")

	return nil
}
