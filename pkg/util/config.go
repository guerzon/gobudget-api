package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBConnString         string        `mapstructure:"DB_CONNSTRING"`
	DBMigrationFiles     string        `mapstructure:"DB_MIGRATION_FILES"`
	AppURL               string        `mapstructure:"APP_URL"`
	ListenAddr           string        `mapstructure:"LISTEN_ADDR"`
	ListenPort           string        `mapstructure:"LISTEN_PORT"`
	Environment          string        `mapstructure:"ENVIRONMENT"`
	SecretKey            string        `mapstructure:"SECRET_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	GmailSenderAddress   string        `mapstructure:"GMAIL_SENDER_ADDRESS"`
	GmailSenderPassword  string        `mapstructure:"GMAIL_SENDER_PASSWORD"`
	MailhogHost          string        `mapstructure:"MAILHOG_HOST"`
	MailhogSenderAddress string        `mapstructure:"MAILHOG_SENDER_ADDRESS"`
}

// viper loads values etiher from app.env or from environment variables
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	// this will allow Viper to prioritize environment variables over config files
	viper.AutomaticEnv()
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	err = viper.Unmarshal(&config)

	return
}
