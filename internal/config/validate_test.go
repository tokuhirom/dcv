package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_UnknownInitialView(t *testing.T) {
	// Setup temp config dir
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create config file with unknown initial_view
	configContent := `[general]
initial_view = "unknown_view"`
	err := os.MkdirAll(filepath.Join(tmpDir, "dcv"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "dcv", "config.toml"), []byte(configContent), 0644)
	require.NoError(t, err)

	// Load should succeed but with unknown value
	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "unknown_view", cfg.General.InitialView)
}
