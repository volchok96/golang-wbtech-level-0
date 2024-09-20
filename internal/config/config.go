package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type NatsStreamingConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	ClusterID string `yaml:"clusterID"`
	ClientID  string `yaml:"clientID"`
}

type AppConfig struct {
	DB DBConfig       `yaml:"db"`
	Nats     NatsStreamingConfig `yaml:"nats"`
}

func GetConfig() (AppConfig, error) {
	var config AppConfig

	file, err := os.ReadFile(".././config.yaml")
	if err != nil {
		log.Error().Err(err).Msg("Failed to read config file")
		return config, err
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal config file")
	}

	return config, nil
}
