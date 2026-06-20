package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/osillation/obs/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConfigEnv(t *testing.T) {
	dir := t.TempDir()
	content := "CLIENT_DOMAIN=test.example.com\nOBS_INGEST_TOKEN=abc123\n# comment\n\nEMPTY=\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.env"), []byte(content), 0644))

	cfg, err := config.ReadEnv(filepath.Join(dir, "config.env"))
	require.NoError(t, err)
	assert.Equal(t, "test.example.com", cfg["CLIENT_DOMAIN"])
	assert.Equal(t, "abc123", cfg["OBS_INGEST_TOKEN"])
	assert.Equal(t, "", cfg["EMPTY"])
	_, ok := cfg["# comment"]
	assert.False(t, ok)
}

func TestWriteConfigEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.env")
	cfg := map[string]string{
		"CLIENT_DOMAIN":    "example.com",
		"OBS_INGEST_TOKEN": "secrettoken",
	}
	require.NoError(t, config.WriteEnv(path, cfg))

	read, err := config.ReadEnv(path)
	require.NoError(t, err)
	assert.Equal(t, "example.com", read["CLIENT_DOMAIN"])
	assert.Equal(t, "secrettoken", read["OBS_INGEST_TOKEN"])
}

func TestClientDirExists(t *testing.T) {
	dir := t.TempDir()
	assert.False(t, config.ClientDirExists(filepath.Join(dir, "nonexistent")))
	assert.True(t, config.ClientDirExists(dir))
}
