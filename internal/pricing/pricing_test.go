package pricing_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

func in(n int) usage.Usage { return usage.Usage{InputFresh: n} }

func TestNewLoader_DefaultsLoaded(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	res := l.ComputeCost("claude-sonnet-4-6", in(1_000_000))
	require.True(t, res.Known)
	assert.InDelta(t, 3.0, res.Cost, 0.001)
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

	res := l.ComputeCost("claude-sonnet-4-6", in(1_000_000))
	require.True(t, res.Known)
	assert.InDelta(t, 1.0, res.Cost, 0.001)
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

	res := l.ComputeCost("gpt-4o", in(1_000_000))
	require.True(t, res.Known)
	assert.InDelta(t, 2.5, res.Cost, 0.001)
}

func TestNewLoader_NoOverrideFile(t *testing.T) {
	l, err := pricing.NewLoader(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	require.NoError(t, err)

	res := l.ComputeCost("gpt-4o", usage.Usage{InputFresh: 1000, Output: 500})
	assert.True(t, res.Known)
}

func TestComputeCost_UnknownModel(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	res := l.ComputeCost("unknown-model", usage.Usage{InputFresh: 1000, Output: 500})
	assert.False(t, res.Known)
	assert.InDelta(t, 0.0, res.Cost, 0.0)
}

func TestComputeCost_InOutSeparate(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	res := l.ComputeCost("claude-sonnet-4-6", usage.Usage{InputFresh: 1_000_000, Output: 1_000_000})
	require.True(t, res.Known)
	assert.InDelta(t, 18.0, res.Cost, 0.001)
}

// The worked example from the change: a real Opus session of 44.6M tokens,
// which the two-bucket model priced at $227.56 and the estimate at ~$0.50.
func TestComputeCost_AllFourTiers(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	res := l.ComputeCost("claude-opus-5", usage.Usage{
		InputFresh: 609,
		CacheWrite: 477_012,
		CacheRead:  43_907_365,
		Output:     225_365,
	})

	require.True(t, res.Known)
	assert.False(t, res.Incomplete())
	assert.InDelta(t, 30.57, res.Cost, 0.01)
}

// Reasoning is a subset of Output. If it were ever added on top, every cost for
// a reasoning model would silently inflate, and no total-only test would catch
// it.
func TestComputeCost_ReasoningIsNotBilledTwice(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	withReasoning := l.ComputeCost("claude-sonnet-4-6", usage.Usage{Output: 1000, Reasoning: 400})
	without := l.ComputeCost("claude-sonnet-4-6", usage.Usage{Output: 1000, Reasoning: 0})

	require.True(t, withReasoning.Known)
	assert.InDelta(t, without.Cost, withReasoning.Cost, 0.0)
}

func TestComputeCost_MissingCacheRateIsExcludedAndReported(t *testing.T) {
	dir := t.TempDir()
	override := filepath.Join(dir, "pricing.yaml")

	require.NoError(t, os.WriteFile(override, []byte(`
models:
  bare-model:
    input_per_million: 10.00
    output_per_million: 20.00
`), 0o600))

	l, err := pricing.NewLoader(override)
	require.NoError(t, err)

	res := l.ComputeCost("bare-model", usage.Usage{InputFresh: 1_000_000, CacheRead: 5_000_000})

	require.True(t, res.Known)
	assert.True(t, res.Incomplete())
	assert.Contains(t, res.MissingTiers, "cache_read_per_million")
	// The 5M cache read tokens contribute nothing rather than being charged at
	// the input rate, which would have added $50.
	assert.InDelta(t, 10.0, res.Cost, 0.001)
}

func TestComputeCost_CompleteRatesDoNotWarn(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	res := l.ComputeCost("claude-opus-5", usage.Usage{CacheRead: 1_000_000, CacheWrite: 1_000_000})

	require.True(t, res.Known)
	assert.False(t, res.Incomplete())
	assert.Empty(t, res.MissingTiers)
}

func TestCacheSavings(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	// 43,907,365 cache reads at $0.50/M cost $21.95; as fresh input at $5.00/M
	// they would have cost $219.54.
	saved, ok := l.CacheSavings("claude-opus-5", usage.Usage{CacheRead: 43_907_365})

	require.True(t, ok)
	assert.InDelta(t, 197.58, saved, 0.01)
}

func TestCacheSavings_NoCachingMeansNoSaving(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	saved, ok := l.CacheSavings("claude-opus-5", usage.Usage{InputFresh: 1000, Output: 500})

	require.True(t, ok)
	assert.InDelta(t, 0.0, saved, 0.0)
}

func TestDefaults_EveryModelDeclaresInputAndOutput(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	all := l.All()
	require.NotEmpty(t, all)

	for model, price := range all {
		assert.Positive(t, price.InputPerMillion, "model %s has no input rate", model)
		assert.Positive(t, price.OutputPerMillion, "model %s has no output rate", model)
	}
}

// claude-opus-5 was missing from the defaults entirely, so every session run on
// it fell into the unknown-model path and reported no cost at all.
func TestDefaults_IncludeClaudeOpus5(t *testing.T) {
	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	price, ok := l.All()["claude-opus-5"]

	require.True(t, ok)
	assert.InDelta(t, 5.0, price.InputPerMillion, 0.001)
	assert.InDelta(t, 25.0, price.OutputPerMillion, 0.001)
	assert.InDelta(t, 0.50, price.CacheReadPerMillion, 0.001)
	assert.InDelta(t, 6.25, price.CacheWritePerMillion, 0.001)
}

func TestSetOverride_WritesAndUpdates(t *testing.T) {
	dir := t.TempDir()
	override := filepath.Join(dir, "pricing.yaml")

	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	err = l.SetOverride(override, "my-model", pricing.ModelPrice{
		InputPerMillion:  0.10,
		OutputPerMillion: 0.20,
	})
	require.NoError(t, err)

	res := l.ComputeCost("my-model", usage.Usage{InputFresh: 1_000_000, Output: 1_000_000})
	require.True(t, res.Known)
	assert.InDelta(t, 0.30, res.Cost, 0.001)

	data, err := os.ReadFile(override)
	require.NoError(t, err)
	assert.Contains(t, string(data), "my-model")
}

func TestSetOverride_WithCacheRates(t *testing.T) {
	dir := t.TempDir()
	override := filepath.Join(dir, "pricing.yaml")

	l, err := pricing.NewLoader("")
	require.NoError(t, err)

	require.NoError(t, l.SetOverride(override, "my-model", pricing.ModelPrice{
		InputPerMillion:      1.00,
		OutputPerMillion:     2.00,
		CacheReadPerMillion:  0.10,
		CacheWritePerMillion: 1.25,
	}))

	res := l.ComputeCost("my-model", usage.Usage{CacheRead: 1_000_000, CacheWrite: 1_000_000})

	require.True(t, res.Known)
	assert.False(t, res.Incomplete())
	assert.InDelta(t, 1.35, res.Cost, 0.001)
}
