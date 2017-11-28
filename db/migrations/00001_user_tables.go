package migrations

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pressly/goose"

	"github.com/iReflect/reflect-app/apps/user/models"
)

func init() {
	goose.AddMigration(Up00001, Down00001)
}

func Up00001(tx *sql.Tx) error {
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}
	gormdb.CreateTable(&models.Role{}, &models.User{}, &models.Team{}, &models.UserTeamAssociation{})

	gormdb.Model(&models.User{}).AddForeignKey("role_id", "roles(id)", "RESTRICT", "RESTRICT")

	gormdb.Model(&models.UserTeamAssociation{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.UserTeamAssociation{}).AddForeignKey("team_id", "teams(id)", "RESTRICT", "RESTRICT")
	gormdb.Model(&models.UserTeamAssociation{}).AddUniqueIndex("unique_user_team", "user_id", "team_id")

	return nil
}

func Down00001(tx *sql.Tx) error {

	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	gormdb.Model(&models.UserTeamAssociation{}).RemoveIndex("unique_user_team");
	gormdb.Model(&models.UserTeamAssociation{}).RemoveForeignKey("team_id", "teams(id)")
	gormdb.Model(&models.UserTeamAssociation{}).RemoveForeignKey("user_id", "users(id)")

	gormdb.Model(&models.User{}).RemoveForeignKey("role_id", "roles(id)")

	gormdb.DropTable(&models.UserTeamAssociation{}, &models.User{}, &models.Team{}, &models.Role{})

	return nil
}
