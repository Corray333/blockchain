package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	PortHTTP int    `yaml:"porthttp"`
	PortP2P  int    `yaml:"portp2p"`
	level    string `yaml:"level"`
}

func LoadConfig() (*Config, error) {

	if err := godotenv.Load("../../.env"); err != nil {
		return nil, fmt.Errorf("error while loading config%s", err.Error())
	}
	var config Config
	configFile, err := os.ReadFile("../../configs/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("error while loading config: %s", err.Error())
	}
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		return nil, fmt.Errorf("error while loading config: %s", err.Error())
	}
	// TODO: config logger
	return &config, nil
}
