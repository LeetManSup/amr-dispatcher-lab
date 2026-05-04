package config

import (
	"encoding/json"
	"os"
)

type AppConfig struct {
	HTTPPort            int    `json:"httpPort"`
	SQLitePath          string `json:"sqlitePath"`
	DefaultMapPath      string `json:"defaultMapPath"`
	DefaultScenarioPath string `json:"defaultScenarioPath"`
	TickDurationMillis  int    `json:"tickDurationMillis"`
	LogLevel            string `json:"logLevel"`
}

func LoadAppConfig(path string) (AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}
	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, err
	}
	return cfg, nil
}
