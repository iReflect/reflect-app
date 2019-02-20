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
	gormDB, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	gormDB.CreateTable(&models.OTP{})

	gormDB.Model(&models.OTP{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")

	return nil
}

// Down00031 ...
func Down00031(tx *sql.Tx) error {

	gormDB, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormDB.Model(&models.OTP{}).RemoveForeignKey("user_id", "users(id)")

	gormDB.DropTable(&models.OTP{})

	return nil
}
