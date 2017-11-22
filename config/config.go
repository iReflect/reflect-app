package config

import "github.com/caarlos0/env"
import "log"

type Config struct {
	DB *DBConfig
}

type DBConfig struct {
	Driver        string `env:"DB_DRIVER"  envDefault:"postgres"`
	DSN           string `env:"DB_DSN"  envDefault:"host=localhost user=ireflect password=1Reflect dbname=ireflect-dev sslmode=disable"`
	MigrationsDir string `env:"MIGRATION_DIR"  envDefault:"db/migrations"`
}

func GetConfig() *Config {
	dbConfig := new(DBConfig)
	env.Parse(dbConfig)
	log.Println("DB::")
	log.Println(dbConfig)
	return &Config{
		DB: dbConfig,
	}
}
