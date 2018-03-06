package config

import (
	"github.com/caarlos0/env"
	"log"
	"os"
)

type Config struct {
	DB          *dbConfig
	Auth        *authConfig
	Redis       *redisConfig
	TimeTracker *timeTrackerConfig
}

var config Config

func init() {
	dbConf := new(dbConfig)
	authConf := new(authConfig)
	redisConf := new(redisConfig)
	timeTrackerConf := new(timeTrackerConfig)
	env.Parse(dbConf)
	env.Parse(authConf)
	env.Parse(redisConf)
	env.Parse(timeTrackerConf)

	googleAppCredential := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if len(googleAppCredential) == 0 {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "config/application_default_credentials.json")

	}
	log.Println("DB::")
	log.Println(dbConf)
	log.Println("Auth::")
	log.Println(authConf)
	log.Println("Redis::")
	log.Println(redisConf)
	log.Println("TimeTracker::")
	log.Println(timeTrackerConf)

	config = Config{
		DB:          dbConf,
		Auth:        authConf,
		Redis:       redisConf,
		TimeTracker: timeTrackerConf,
	}
}

type dbConfig struct {
	Driver        string `env:"DB_DRIVER"  envDefault:"postgres"`
	DSN           string `env:"DB_DSN"  envDefault:"host=localhost user=ireflect password=1Reflect dbname=ireflect-dev sslmode=disable"`
	MigrationsDir string `env:"MIGRATION_DIR"  envDefault:"db/migrations"`
	LogEnabled    bool   `env:"DB_LOG_ENABLED"  envDefault:"true"`
}

type authConfig struct {
	Secret string `env:"AUTH_SECRET"  envDefault:"secret"`
}

type redisConfig struct {
	Address string `env:"REDIS_ADDRESS"  envDefault:":6379"`
}

type timeTrackerConfig struct {
	ScriptID          string `env:"TIMETRACKER_SCRIPT_ID"  envDefault:"MBPTr9ro72YqzPNl1DkDD9ldaih63P1hV"`
	FnGetTimeLog      string `env:"TIMETRACKER_FN_GETTIMELOG"  envDefault:"GetProjectTimeLogs"`
	GoogleCredentials string `env:"TIMETRACKER_CREDENTIALS"  envDefault:"config/timetracker_credentials.json"`
}

// GetConfig ...
func GetConfig() *Config {
	return &config
}
