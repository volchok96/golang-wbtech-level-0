package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Kafka struct {
		Broker  string `yaml:"broker"`
		GroupID string `yaml:"group_id"`
		Topic   string `yaml:"topic"`
	}
	Postgres struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	}
	Memcached struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	}
}

func GetConfig() (AppConfig, error) {
	var config AppConfig

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	configFileName := fmt.Sprintf(".././config.%s.yaml", env)

	file, err := os.ReadFile(configFileName)
	if err != nil {
		return config, fmt.Errorf("error reading config file %s: %w", configFileName, err)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return config, fmt.Errorf("error parsing config file %s: %w", configFileName, err)
	}

	return config, nil
}
