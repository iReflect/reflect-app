package migrations

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00006, Down00006)
}

// Up00006 ...
func Up00006(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.AutoMigrate(&models.SprintMember{})

	gormdb.Model(&models.SprintMember{}).AddForeignKey("sprint_id", "sprints(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.SprintMember{}).AddForeignKey("member_id", "users(id)", "RESTRICT", "RESTRICT")

	gormdb.Model(&models.SprintMember{}).AddIndex("index_sprint_members_sprint_id_member_id", "sprint_id", "member_id")

	return nil
}

// Down00006 ...
func Down00006(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.SprintMember{}).RemoveIndex("index_sprint_members_sprint_id_member_id")

	gormdb.Model(&models.SprintMember{}).RemoveForeignKey("sprint_id", "sprints(id)")
	gormdb.Model(&models.SprintMember{}).RemoveForeignKey("member_id", "users(id)")

	gormdb.DropTable(&models.SprintMember{})

	return nil
}
