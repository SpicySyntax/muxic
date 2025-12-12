package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	DefaultDevice *string `json:"default_device,omitempty"`
}

const configFileName = "muxic_config.json"

func LoadConfig() (Config, error) {
	var config Config
	data, err := os.ReadFile(configFileName)
	if os.IsNotExist(err) {
		return config, nil
	}
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

func (c Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFileName, data, 0644)
}
