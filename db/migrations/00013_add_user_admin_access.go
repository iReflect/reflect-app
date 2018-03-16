package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"
)

//Define only the fields used in this migration and not full model.
type UserAdminAccess struct {
	gorm.Model
	IsAdmin bool `gorm:"default:false; not null"`
}

func (UserAdminAccess) TableName() string {
	return "users"
}

func init() {
	goose.AddMigration(Up00013, Down00013)
}

// Up00013 ...
func Up00013(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	// Add a column
	gormdb.AutoMigrate(&UserAdminAccess{})

	return nil
}

// Down00011 ...
func Down00013(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	// Drop a column
	gormdb.Model(&UserAdminAccess{}).DropColumn("is_admin")

	return nil
}
