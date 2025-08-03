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

func TestLoad_FromCurrentDir(t *testing.T) {
	// Create temp dir and make it current dir
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create config file
	configContent := `[general]
initial_view = "compose"`
	err = os.WriteFile("dcv.toml", []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "compose", cfg.General.InitialView)
}

func TestLoad_FromHomeDir(t *testing.T) {
	// Setup temp home
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create config file
	configContent := `[general]
initial_view = "projects"`
	err := os.WriteFile(filepath.Join(tmpDir, ".dcv.toml"), []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "projects", cfg.General.InitialView)
}

func TestLoad_InvalidTOML(t *testing.T) {
	// Create temp dir and make it current dir
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldCwd) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create invalid config file
	err = os.WriteFile("dcv.toml", []byte("invalid toml content"), 0644)
	require.NoError(t, err)

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestSave(t *testing.T) {
	// Setup temp config dir
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := &Config{
		General: GeneralConfig{
			InitialView: "projects",
		},
	}

	err := cfg.Save()
	require.NoError(t, err)

	// Check file was created
	expectedPath := filepath.Join(tmpDir, "dcv", "config.toml")
	assert.FileExists(t, expectedPath)

	// Load it back
	loadedCfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "projects", loadedCfg.General.InitialView)
}
