package config

import (
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	DB   *DBConfig
	Auth *AuthConfig
}

type DBConfig struct {
	Driver        string `env:"DB_DRIVER"  envDefault:"postgres"`
	DSN           string `env:"DB_DSN"  envDefault:"host=localhost user=ireflect password=1Reflect dbname=ireflect-dev sslmode=disable"`
	MigrationsDir string `env:"MIGRATION_DIR"  envDefault:"db/migrations"`
	LogEnabled    bool   `env:"DB_LOG_ENABLED"  envDefault:"false"`
}

type AuthConfig struct {
	Secret string `env:"AUTH_SECRET"  envDefault:"secret"`
}

func GetConfig() *Config {

	dbConfig := new(DBConfig)
	authConfig := new(AuthConfig)
	env.Parse(dbConfig)
	env.Parse(authConfig)
	log.Println("DB::")
	log.Println(dbConfig)
	log.Println("Auth::")
	log.Println(authConfig)
	return &Config{
		DB:   dbConfig,
		Auth: authConfig,
	}
}
