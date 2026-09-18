package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

var (
	PORT         string
	ENV          string
	DATABASE_URL string
	REDIS_URL    string
	NAME         string
)

func Load(env string) {
	slog.Info("loading config", "env", env)

	viper.SetDefault("ENV", "develpment")
	viper.SetDefault("PORT", "3000")
	viper.SetDefault("NAME", "module_transfer")
	viper.SetDefault("DATABASE_URL", "postgres://postgres:potgres@localhost:5433/goproj")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")

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
}

func IsDevelopment() bool { return ENV == "development" }
func IsProduction() bool  { return ENV == "production" }
func IsTest() bool        { return ENV == "test" }
