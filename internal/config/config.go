package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type StravaConfig struct {
	Email    string
	Password string
}

func ParseConfig(path string) (*StravaConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open config file '%s': %w", path, err)
	}
	defer file.Close()
	body, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read from config file '%s': %w", path, err)
	}
	var config StravaConfig
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal json from config file '%s': %w", path, err)
	}
	if config.Email == "" {
		return nil, fmt.Errorf("mandatory field 'Email' not found in '%s'", path)
	}
	if config.Password == "" {
		return nil, fmt.Errorf("mandatory field 'Password' not found in '%s'", path)
	}
	return &config, nil
}
