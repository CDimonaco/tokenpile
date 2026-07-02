package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cdimonaco/tokenpile/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_UsesEnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TOKENPILE_CONFIG_DIR", dir)
	t.Setenv("TOKENPILE_DATA_DIR", dir)

	p := config.Resolve()

	assert.Equal(t, dir, p.ConfigDir)
	assert.Equal(t, dir, p.DataDir)
	assert.Equal(t, filepath.Join(dir, "tokenpile.db"), p.DBPath)
	assert.Equal(t, filepath.Join(dir, "pricing.yaml"), p.PricingOverride)
	assert.Equal(t, filepath.Join(dir, "identity.key"), p.IdentityKeyPath)
	assert.Equal(t, filepath.Join(dir, "identity.pub"), p.IdentityPubPath)
	assert.Equal(t, filepath.Join(dir, "credentials"), p.CredentialsPath)
}

func TestResolve_UsesXDGConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TOKENPILE_CONFIG_DIR", "")
	t.Setenv("TOKENPILE_DATA_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)

	p := config.Resolve()

	assert.Equal(t, filepath.Join(dir, "tokenpile"), p.ConfigDir)
	assert.Equal(t, filepath.Join(dir, "tokenpile"), p.DataDir)
}

func TestEnsureDirs_CreatesDirs(t *testing.T) {
	base := t.TempDir()
	p := config.Paths{
		ConfigDir: filepath.Join(base, "cfg"),
		DataDir:   filepath.Join(base, "data"),
	}

	err := config.EnsureDirs(p)
	require.NoError(t, err)

	info, err := os.Stat(p.ConfigDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	info, err = os.Stat(p.DataDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestEnsureDirs_Idempotent(t *testing.T) {
	base := t.TempDir()
	p := config.Paths{
		ConfigDir: filepath.Join(base, "cfg"),
		DataDir:   filepath.Join(base, "data"),
	}

	require.NoError(t, config.EnsureDirs(p))
	require.NoError(t, config.EnsureDirs(p))
}
