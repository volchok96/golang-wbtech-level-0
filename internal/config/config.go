package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Kafka struct {
		Broker string
		GroupID string
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

