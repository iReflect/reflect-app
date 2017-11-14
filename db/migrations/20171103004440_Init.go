package migrations

import (
	"database/sql"

	"github.com/iReflect/reflect-app/db/migrations/base"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(Up00001, Down00001)
}

func Up00001(tx *sql.Tx) error {
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.CreateTable(&base.User{})
	gormdb.CreateTable(&base.Group{})
	gormdb.CreateTable(&base.Role{})
	gormdb.CreateTable(&base.Category{})
	gormdb.CreateTable(&base.Item{})
	gormdb.CreateTable(&base.ItemType{})
	return nil
}

func Down00001(tx *sql.Tx) error {

	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	gormdb.DropTable(&base.Role{})
	gormdb.DropTable(&base.User{})
	gormdb.DropTable(&base.Group{})
	gormdb.DropTable(&base.ItemType{})
	gormdb.DropTable(&base.Category{})
	gormdb.DropTable(&base.Item{})

	return nil
}
