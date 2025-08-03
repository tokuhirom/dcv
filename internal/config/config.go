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
	// Get config file paths
	configPaths := getConfigPaths()

	// Start with default config
	cfg := Default()

	// Try to load from each path
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			// File exists, try to load it
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
			}

			if err := toml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
			}

			// Successfully loaded, return
			return cfg, nil
		}
	}

	// No config file found, return default
	return cfg, nil
}

// getConfigPaths returns the paths where config files are searched
func getConfigPaths() []string {
	var paths []string

	// 1. Current directory
	paths = append(paths, "dcv.toml")

	// 2. User config directory
	if configDir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(configDir, "dcv", "config.toml"))
	}

	// 3. Home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".dcv.toml"))
	}

	return paths
}

// Save saves the configuration to the user config directory
func (c *Config) Save() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config directory: %w", err)
	}

	dcvConfigDir := filepath.Join(configDir, "dcv")
	if err := os.MkdirAll(dcvConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(dcvConfigDir, "config.toml")

	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
