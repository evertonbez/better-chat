package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/spf13/viper"
)

const DevAuthSecret = "dev-insecure-auth-secret-change-me"

var (
	PORT         string
	ENV          string
	DATABASE_URL string
	REDIS_URL    string
	NAME         string

	AUTH_SECRET string

	SESSION_TTL time.Duration

	SESSION_UPDATE_AGE time.Duration

	SESSION_CACHE_TTL time.Duration

	COOKIE_NAME   string
	COOKIE_DOMAIN string
	COOKIE_SECURE bool
)

func Load(env string) {
	slog.Info("loading config", "env", env)

	viper.SetDefault("ENV", "development")
	viper.SetDefault("PORT", "3000")
	viper.SetDefault("NAME", "module_api")
	viper.SetDefault("DATABASE_URL", "postgres://postgres:potgres@localhost:5433/goproj")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")

	viper.SetDefault("AUTH_SECRET", DevAuthSecret)
	viper.SetDefault("SESSION_TTL", "168h")
	viper.SetDefault("SESSION_UPDATE_AGE", "24h")
	viper.SetDefault("SESSION_CACHE_TTL", "5m")

	viper.SetDefault("COOKIE_NAME", "better-chat.session_token")
	viper.SetDefault("COOKIE_DOMAIN", "")
	viper.SetDefault("COOKIE_SECURE", false)

	viper.AutomaticEnv()

	if env == "" {
		viper.SetConfigFile(".env")
		if err := viper.ReadInConfig(); err != nil {
			slog.Warn("no .env file loaded, falling back to defaults and enviroment", "error", err)
		}
	}

	PORT = viper.GetString("PORT")
	ENV = viper.GetString("ENV")
	DATABASE_URL = viper.GetString("DATABASE_URL")
	REDIS_URL = viper.GetString("REDIS_URL")
	NAME = viper.GetString("NAME")

	AUTH_SECRET = viper.GetString("AUTH_SECRET")
	SESSION_TTL = viper.GetDuration("SESSION_TTL")
	SESSION_UPDATE_AGE = viper.GetDuration("SESSION_UPDATE_AGE")
	SESSION_CACHE_TTL = viper.GetDuration("SESSION_CACHE_TTL")

	COOKIE_NAME = viper.GetString("COOKIE_NAME")
	COOKIE_DOMAIN = viper.GetString("COOKIE_DOMAIN")
	COOKIE_SECURE = viper.GetBool("COOKIE_SECURE")

	if IsProduction() {

		if AUTH_SECRET == DevAuthSecret || AUTH_SECRET == "" {
			slog.Error("AUTH_SECRET must be set in production")
			os.Exit(1)
		}
		if !COOKIE_SECURE {
			slog.Warn("COOKIE_SECURE is off in production, session cookies will travel over plain HTTP")
		}
	}

	if SESSION_UPDATE_AGE >= SESSION_TTL {
		slog.Warn("SESSION_UPDATE_AGE >= SESSION_TTL, sessions will never slide",
			"update_age", SESSION_UPDATE_AGE, "ttl", SESSION_TTL)
	}
}

func IsDevelopment() bool { return ENV == "development" }
func IsProduction() bool  { return ENV == "production" }
func IsTest() bool        { return ENV == "test" }
