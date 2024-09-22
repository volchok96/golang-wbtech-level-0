package config

import (
	"fmt"
	"os"
	"wb-kafka-service/pkg/logger"
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

func GetConfig(log *logger.Logger) (AppConfig, error) {
	var config AppConfig

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	configFileName := fmt.Sprintf(".././config.%s.yaml", env)

	file, err := os.ReadFile(configFileName)
	if err != nil {
		log.Error(fmt.Sprintf("Error reading config file %s", configFileName), err)
		return config, fmt.Errorf("error reading config file %s: %w", configFileName, err)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Error(fmt.Sprintf("Error parsing config file %s", configFileName), err)
		return config, fmt.Errorf("error parsing config file %s: %w", configFileName, err)
	}

	log.Info(fmt.Sprintf("Successfully loaded config from %s", configFileName))
	return config, nil
}
