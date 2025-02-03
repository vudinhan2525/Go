package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DbDriver             string        `mapstructure:"DBDRIVER"`
	DbSource             string        `mapstructure:"DBSOURCE"`
	APIEndpoint          string        `mapstructure:"API_ENDPOINT"`
	GrpcAPIEndpoint      string        `mapstructure:"GRPC_API_ENDPOINT"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	RedisAddress         string        `mapstructure:"REDIS_SERVER_ADDRESS"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
	TokenDuration        time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
