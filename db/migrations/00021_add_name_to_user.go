package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

//Define only the fields used in this migration and not full model.
//type Category struct {
//	gorm.Model
//	Weight int
//}

func init() {
	goose.AddMigration(Up00021, Down00021)
}

// Up00021 ...
func Up00021(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type user struct {
		Name string `gorm:"type:varchar(255); not null;default:''"`
	}
	gormdb.Exec("UPDATE users SET name = CONCAT(first_name, '', last_name);")
	gormdb.AutoMigrate(&user{})
	return nil
}

// Down00021 ...
func Down00021(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	type user struct {
		Name string `gorm:"type:varchar(255); not null"`
	}

	// Drop a column
	gormdb.Model(&user{}).DropColumn("name")

	return nil
}
