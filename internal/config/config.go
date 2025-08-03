package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	// General settings
	General GeneralConfig `toml:"general"`
}

// GeneralConfig contains general application settings
type GeneralConfig struct {
	// InitialView specifies which view to show on startup
	// Valid values: "compose", "docker", "projects"
	InitialView string `toml:"initial_view"`
}

// Default returns the default configuration
func Default() *Config {
	return &Config{
		General: GeneralConfig{
			InitialView: "docker", // Default to docker container list
		},
	}
}

// Load loads configuration from file
func Load() (*Config, error) {
	// Get config file path
	configPath := getConfigPath()

	// Start with default config
	cfg := Default()

	// If we couldn't determine config path, return default
	if configPath == "" {
		return cfg, nil
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); err != nil {
		// File doesn't exist, return default
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse config file
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return cfg, nil
}

// getConfigPath returns the path where config file is located
func getConfigPath() string {
	// User config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}

	return filepath.Join(configDir, "dcv", "config.toml")
}
