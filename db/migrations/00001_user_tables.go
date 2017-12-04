package migrations

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	models "github.com/iReflect/reflect-app/db/base_models"
)

func init() {
	goose.AddMigration(Up00001, Down00001)
}

func Up00001(tx *sql.Tx) error {
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	gormdb.CreateTable(&models.Role{}, &models.User{}, &models.Team{}, &models.UserTeamAssociation{}, &models.UserProfile{})

	gormdb.Model(&models.UserTeamAssociation{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.UserTeamAssociation{}).AddForeignKey("team_id", "teams(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.UserTeamAssociation{}).AddUniqueIndex("unique_user_team", "user_id", "team_id")

	gormdb.Model(&models.UserProfile{}).AddForeignKey("role_id", "roles(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.UserProfile{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")

	return nil
}

func Down00001(tx *sql.Tx) error {

	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.UserProfile{}).RemoveForeignKey("user_id", "users(id)")
	gormdb.Model(&models.UserProfile{}).RemoveForeignKey("role_id", "roles(id)")

	gormdb.Model(&models.UserTeamAssociation{}).RemoveIndex("unique_user_team");
	gormdb.Model(&models.UserTeamAssociation{}).RemoveForeignKey("team_id", "teams(id)")
	gormdb.Model(&models.UserTeamAssociation{}).RemoveForeignKey("user_id", "users(id)")

	gormdb.DropTable(&models.UserProfile{}, &models.UserTeamAssociation{}, &models.User{}, &models.Team{}, &models.Role{})

	return nil
}
