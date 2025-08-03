package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	assert.NotNil(t, cfg)
	assert.Equal(t, "docker", cfg.General.InitialView)
}

func TestLoad_NoConfigFile(t *testing.T) {
	// Ensure no config files exist in test environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "docker", cfg.General.InitialView)
}


func TestLoad_FromUserConfigDir(t *testing.T) {
	// Setup temp config dir
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create config file
	configContent := `[general]
initial_view = "projects"`
	err := os.MkdirAll(filepath.Join(tmpDir, "dcv"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "dcv", "config.toml"), []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "projects", cfg.General.InitialView)
}

func TestLoad_InvalidTOML(t *testing.T) {
	// Setup temp config dir
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create invalid config file
	err := os.MkdirAll(filepath.Join(tmpDir, "dcv"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "dcv", "config.toml"), []byte("invalid toml content"), 0644)
	require.NoError(t, err)

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

