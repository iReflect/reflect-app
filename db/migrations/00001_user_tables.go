package migrations

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/db/base/models"
)

func init() {
	goose.AddMigration(Up00001, Down00001)
}

// Up00001 Create user tables
func Up00001(tx *sql.Tx) error {
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	gormdb.CreateTable(&models.Role{}, &models.User{}, &models.Team{}, &models.UserTeam{}, &models.UserProfile{})

	gormdb.Model(&models.UserTeam{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.UserTeam{}).AddForeignKey("team_id", "teams(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.UserTeam{}).AddUniqueIndex("unique_user_team", "user_id", "team_id")

	gormdb.Model(&models.UserProfile{}).AddForeignKey("role_id", "roles(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.UserProfile{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")

	return nil
}

// Down00001 drop user tables
func Down00001(tx *sql.Tx) error {

	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.UserProfile{}).RemoveForeignKey("user_id", "users(id)")
	gormdb.Model(&models.UserProfile{}).RemoveForeignKey("role_id", "roles(id)")

	gormdb.Model(&models.UserTeam{}).RemoveIndex("unique_user_team")
	gormdb.Model(&models.UserTeam{}).RemoveForeignKey("team_id", "teams(id)")
	gormdb.Model(&models.UserTeam{}).RemoveForeignKey("user_id", "users(id)")

	gormdb.DropTable(&models.UserProfile{}, &models.UserTeam{}, &models.User{}, &models.Team{}, &models.Role{})

	return nil
}
