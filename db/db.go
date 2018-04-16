package db

import (
	"errors"
	"log"

	"github.com/pressly/goose"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	// Postgres DB wrapper for GORM
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/iReflect/reflect-app/config"
)

// Initialize GORM DB instance
func Initialize(config *config.Config) *gorm.DB {

	db, err := gorm.Open(config.DB.Driver, config.DB.DSN)
	if err != nil {
		log.Fatal("Could not connect database", err)
	}
	db.LogMode(config.DB.LogEnabled)
	return db
}

// Migrate to latest version
func Migrate(config *config.Config, db *gorm.DB) error {

	goose.Up(db.DB(), config.DB.MigrationsDir)

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
