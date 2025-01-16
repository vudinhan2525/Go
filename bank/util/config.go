package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DbDriver          string        `mapstructure:"DBDRIVER"`
	DbSource          string        `mapstructure:"DBSOURCE"`
	APIEndpoint       string        `mapstructure:"API_ENDPOINT"`
	TokenSymmetricKey string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	TokenDuration     time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
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
