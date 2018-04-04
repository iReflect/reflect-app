package db

import (
	"errors"
	"log"

	"github.com/pressly/goose"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/iReflect/reflect-app/config"
)

// DB is the Database instance to be used across the project
var DB *gorm.DB

// Initialize GORM DB instance
func Initialize(config *config.Config) *gorm.DB {
	var err error
	DB, err = gorm.Open(config.DB.Driver, config.DB.DSN)
	if err != nil {
		log.Fatal("Could not connect database", err)
	}
	DB.LogMode(config.DB.LogEnabled)
	return DB
}

// Migrate to latest version
func Migrate(config *config.Config) error {

	goose.Up(DB.DB(), config.DB.MigrationsDir)

	return nil
}

// GetFromContext Get DB from Gin Context
func GetFromContext(c *gin.Context) (*gorm.DB, error) {
	dbi, _ := c.Get("DB")
	db, ok := dbi.(*gorm.DB)
	if !ok {
		return nil, errors.New("No db in context")
	}
	return db, nil
}
