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
	gormdb.CreateTable(&base.User{}, &base.Group{}, &base.Role{})
	gormdb.Model(&base.Role{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&base.Role{}).AddForeignKey("group_id", "groups(id)", "RESTRICT", "RESTRICT")

	gormdb.CreateTable(&base.Category{}, &base.Item{}, &base.ItemType{})
	gormdb.Model(&base.Item{}).AddForeignKey("category_id", "categories(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&base.Item{}).AddForeignKey("item_type_id", "item_types(id)", "RESTRICT", "RESTRICT")

	return nil
}

func Down00001(tx *sql.Tx) error {

	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&base.Role{}).RemoveForeignKey("user_id", "users(id)")
	gormdb.Model(&base.Role{}).RemoveForeignKey("group_id", "groups(id)")
	gormdb.DropTable(&base.User{}, &base.Group{}, &base.Role{})

	gormdb.Model(&base.Item{}).RemoveForeignKey("category_id", "categories(id)")
	gormdb.Model(&base.Item{}).RemoveForeignKey("item_type_id", "item_types(id)")
	gormdb.DropTable(&base.Category{}, &base.Item{}, &base.ItemType{})

	return nil
}
