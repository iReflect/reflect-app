package config

import (
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	DB          *DBConfig
	Auth        *AuthConfig
	Redis       *RedisConfig
	TimeTracker *TimeTrackerConfig
}

var config Config

func init() {
	dbConfig := new(DBConfig)
	authConfig := new(AuthConfig)
	redisConfig := new(RedisConfig)
	timeTrackerConfig := new(TimeTrackerConfig)
	env.Parse(dbConfig)
	env.Parse(authConfig)
	env.Parse(redisConfig)
	env.Parse(timeTrackerConfig)
	log.Println("DB::")
	log.Println(dbConfig)
	log.Println("Auth::")
	log.Println(authConfig)
	log.Println("Redis::")
	log.Println(redisConfig)
	log.Println("TimeTracker::")
	log.Println(timeTrackerConfig)
	config = Config{
		DB:          dbConfig,
		Auth:        authConfig,
		Redis:       redisConfig,
		TimeTracker: timeTrackerConfig,
	}
}

type DBConfig struct {
	Driver        string `env:"DB_DRIVER"  envDefault:"postgres"`
	DSN           string `env:"DB_DSN"  envDefault:"host=localhost user=ireflect password=1Reflect dbname=ireflect-dev sslmode=disable"`
	MigrationsDir string `env:"MIGRATION_DIR"  envDefault:"db/migrations"`
	LogEnabled    bool   `env:"DB_LOG_ENABLED"  envDefault:"true"`
}

type AuthConfig struct {
	Secret string `env:"AUTH_SECRET"  envDefault:"secret"`
}

type RedisConfig struct {
	Address string `env:"REDIS_ADDRESS"  envDefault:":6379"`
}

type TimeTrackerConfig struct {
	ScriptID          string `env:"TIMETRACKER_SCRIPT_ID"  envDefault:"MBPTr9ro72YqzPNl1DkDD9ldaih63P1hV"`
	FnGetTimeLog      string `env:"TIMETRACKER_FN_GETTIMELOG"  envDefault:"GetProjectTimeLogs"`
	GoogleCredentials string `env:"TIMETRACKER_CREDENTIALS"  envDefault:"config/timetracker_credentials.json"`
}

func GetConfig() *Config {
	return &config
}
