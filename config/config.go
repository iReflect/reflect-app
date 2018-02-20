package config

import (
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	DB   *DBConfig
	Auth *AuthConfig
	Redis *RedisConfig
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

type RedisConfig struct {
	Address string `env:"REDIS_ADDRESS"  envDefault:":6379"`
}

func GetConfig() *Config {

	dbConfig := new(DBConfig)
	authConfig := new(AuthConfig)
	redisConfig := new(RedisConfig)
	env.Parse(dbConfig)
	env.Parse(authConfig)
	env.Parse(redisConfig)
	log.Println("DB::")
	log.Println(dbConfig)
	log.Println("Auth::")
	log.Println(authConfig)
	log.Println("Redis::")
	log.Println(redisConfig)
	return &Config{
		DB:   dbConfig,
		Auth: authConfig,
		Redis: redisConfig,
	}
}
