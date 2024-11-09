package configuration

import (
	"context"
	"os"

	"diy.blockchain.org/m/logger"
	"gopkg.in/yaml.v2"
)

type Config struct {
	HttpPort string `yaml:"http_port"`
}

var InstanceConfig Config

func LoadConfig(ctx context.Context, configPath string) {
	if path, ok := os.LookupEnv("CONFIG_PATH"); ok {
		configPath = path
	}
	yamlFile, err := os.ReadFile(configPath)

	if err != nil {
		logger.Errorf("Failed to read config file: %v", err)
	}

	if err := yaml.Unmarshal(yamlFile, &InstanceConfig); err != nil {
		logger.Errorf("Failed to parse config file: %v", err)
	}
}
