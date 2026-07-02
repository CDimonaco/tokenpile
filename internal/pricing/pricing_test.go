package pricing_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader_DefaultsLoaded(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	cost, ok := l.ComputeCost("claude-sonnet-4-6", 1_000_000, 0)
	require.True(t, ok)
	assert.InDelta(t, 3.0, cost, 0.001)
}

func TestNewLoader_UserOverrideTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	override := filepath.Join(dir, "pricing.yaml")

	err := os.WriteFile(override, []byte(`
models:
  claude-sonnet-4-6:
    input_per_million: 1.00
    output_per_million: 5.00
`), 0o600)
	require.NoError(t, err)

	l, err := pricing.NewLoader(override)
	require.NoError(t, err)

	cost, ok := l.ComputeCost("claude-sonnet-4-6", 1_000_000, 0)
	require.True(t, ok)
	assert.InDelta(t, 1.0, cost, 0.001)
}

func TestNewLoader_DefaultFillsMissingOverride(t *testing.T) {
	dir := t.TempDir()
	override := filepath.Join(dir, "pricing.yaml")

	err := os.WriteFile(override, []byte(`
models:
  my-model:
    input_per_million: 0.50
    output_per_million: 1.00
`), 0o600)
	require.NoError(t, err)

	l, err := pricing.NewLoader(override)
	require.NoError(t, err)

	cost, ok := l.ComputeCost("gpt-4o", 1_000_000, 0)
	require.True(t, ok)
	assert.InDelta(t, 2.5, cost, 0.001)
}

func TestNewLoader_NoOverrideFile(t *testing.T) {
	l, err := pricing.NewLoader(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	require.NoError(t, err)

	_, ok := l.ComputeCost("gpt-4o", 1000, 500)
	assert.True(t, ok)
}

func TestComputeCost_UnknownModel(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	cost, ok := l.ComputeCost("unknown-model", 1000, 500)
	assert.False(t, ok)
	assert.Equal(t, 0.0, cost)
}

func TestComputeCost_InOutSeparate(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	cost, ok := l.ComputeCost("claude-sonnet-4-6", 1_000_000, 1_000_000)
	require.True(t, ok)
	assert.InDelta(t, 18.0, cost, 0.001)
}

func TestSetOverride_WritesAndUpdates(t *testing.T) {
	dir := t.TempDir()
	override := filepath.Join(dir, "pricing.yaml")

	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	err = l.SetOverride(override, "my-model", 0.10, 0.20)
	require.NoError(t, err)

	cost, ok := l.ComputeCost("my-model", 1_000_000, 1_000_000)
	require.True(t, ok)
	assert.InDelta(t, 0.30, cost, 0.001)

	data, err := os.ReadFile(override)
	require.NoError(t, err)
	assert.Contains(t, string(data), "my-model")
}
