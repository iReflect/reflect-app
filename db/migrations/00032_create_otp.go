package migrations

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00032, Down00032)
}

// Up00032 ...
func Up00032(tx *sql.Tx) error {
	gormDB, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	gormDB.CreateTable(&models.OTP{})

	gormDB.Model(&models.OTP{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")

	return nil
}

// Down00032 ...
func Down00032(tx *sql.Tx) error {

	gormDB, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormDB.Model(&models.OTP{}).RemoveForeignKey("user_id", "users(id)")

	gormDB.DropTable(&models.OTP{})

	return nil
}
