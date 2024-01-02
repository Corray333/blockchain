package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	PortHTTP  int      `yaml:"porthttp"`
	PortP2P   int      `yaml:"portp2p"`
	env       string   `yaml:"env"`
	BootNodes []string `yaml:"boot_nodes"`
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

	var log *slog.Logger

	switch config.env {
	case "debug":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		file, err := os.Create(fmt.Sprintf("../../logs/%s.log", time.Now().Format("01-02-2006-15:04:05")))
		if err != nil {
			return nil, fmt.Errorf("error while creating log file: %s", err.Error())
		}
		defer file.Close()
		log = slog.New(
			slog.NewJSONHandler(file, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	slog.SetDefault(log)

	return &config, nil
}
