package migrations

import (
	"database/sql"

	"github.com/iReflect/reflect-app/db/base/models"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00030, Down00030)
}

// Up00030 ...
func Up00030(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type User struct {
		Password string `gorm:"type:varchar(255)"`
	}

	gormdb.AutoMigrate(&User{})

	return nil
}

// Down00030 ...
func Down00030(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.User{}).DropColumn("password")

	return nil
}
