package config

import (
	// "github.com/spf13/viper"
	"os"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Kafka struct {
		Broker string
		Topic  string
	}
	Postgres struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
	}
}

// func GetConfig() (AppConfig, error) {
// 	viper.SetConfigName("config")
// 	viper.SetConfigType("yaml")
// 	viper.AddConfigPath(".")

// 	if err := viper.ReadInConfig(); err != nil {
// 		return AppConfig{}, err
// 	}

// 	var config AppConfig
// 	if err := viper.Unmarshal(&config); err != nil {
// 		return AppConfig{}, err
// 	}

// 	return config, nil
// }

func GetConfig() (AppConfig, error) {
	var config AppConfig

	file, err := os.ReadFile(".././config.yaml")

	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		//log.Println(err)
	}
	return config, err
}

