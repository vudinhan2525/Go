package util

import "github.com/spf13/viper"

type Config struct {
	DbDriver    string `mapstructure:"DBDRIVER"`
	DbSource    string `mapstructure:"DBSOURCE"`
	APIEndpoint string `mapstructure:"API_ENDPOINT"`
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
