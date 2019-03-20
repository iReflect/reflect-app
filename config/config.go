package config

import (
	"log"
	"os"

	"github.com/caarlos0/env"
)

// Config ...
type Config struct {
	DB          *dbConfig
	Server      *serverConfig
	Redis       *redisConfig
	TimeTracker *timeTrackerConfig
	Email       *emailConfig
}

var config Config

func init() {
	dbConf := new(dbConfig)
	serverConf := new(serverConfig)
	redisConf := new(redisConfig)
	timeTrackerConf := new(timeTrackerConfig)
	emailConfig := new(emailConfig)
	env.Parse(dbConf)
	env.Parse(serverConf)
	env.Parse(redisConf)
	env.Parse(timeTrackerConf)
	env.Parse(emailConfig)
	googleAppCredential := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if len(googleAppCredential) == 0 {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "config/application_default_credentials.json")

	}
	log.Println("DB::")
	log.Println(dbConf)
	log.Println("Auth::")
	log.Println(serverConf)
	log.Println("Redis::")
	log.Println(redisConf)
	log.Println("TimeTracker::")
	log.Println(timeTrackerConf)
	log.Println("Email::")
	log.Println(emailConfig)

	config = Config{
		DB:          dbConf,
		Server:      serverConf,
		Redis:       redisConf,
		TimeTracker: timeTrackerConf,
		Email:       emailConfig,
	}
}

type dbConfig struct {
	Driver        string `env:"DB_DRIVER"  envDefault:"postgres"`
	DSN           string `env:"DB_DSN"  envDefault:"host=localhost user=ireflect password=1Reflect dbname=ireflect_dev sslmode=disable"`
	MigrationsDir string `env:"MIGRATION_DIR"  envDefault:"db/migrations"`
	LogEnabled    bool   `env:"DB_LOG_ENABLED"  envDefault:"true"`
}

type serverConfig struct {
	SessionSecret      string   `env:"SESSION_SECRET"  envDefault:"secret"`
	SessionAge         int      `env:"SESSION_AGE"  envDefault:"604800"` // 1 week
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:4200,http://localhost:3000"`
	LoginURL           string   `env:"LOGIN_URL" envDefault:"http://localhost:4200/login"`
	EncryptionKey      string   `env:"ENCRYPTION_KEY" envDefault:"DUMMY_KEY__FOR_LOCAL_DEV"`
	TimeZone           string   `env:"TIME_ZONE"  envDefault:"Asia/Kolkata"`
}

type redisConfig struct {
	Address string `env:"REDIS_ADDRESS"  envDefault:":6379"`
}

type timeTrackerConfig struct {
	ScriptID          string `env:"TIMETRACKER_SCRIPT_ID"  envDefault:"MBPTr9ro72YqzPNl1DkDD9ldaih63P1hV"`
	FnGetTimeLog      string `env:"TIMETRACKER_FN_GETTIMELOG"  envDefault:"GetProjectTimeLogs"`
	GoogleCredentials string `env:"TIMETRACKER_CREDENTIALS"  envDefault:"config/timetracker_credentials.json"`
	TimeZone          string `env:"TIMETRACKER_TIME_ZONE"  envDefault:"Asia/Kolkata"`
}

type emailConfig struct {
	Username string `env:"EMAIL_USERNAME" envDefault:""`
	Password string `env:"EMAIL_PASSWORD" envDefault:""`
	Host     string `env:"EMAIL_HOST" envDefault:""`
	Port     string `env:"EMAIL_PORT" envDefault:""`
}

// GetConfig ...
func GetConfig() *Config {
	return &config
}
