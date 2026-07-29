package pricing

import (
	_ "embed"
	"fmt"
	"log/slog"
	"maps"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/cdimonaco/tokenpile/internal/usage"
)

//go:embed pricing.defaults.yaml
var defaultsYAML []byte

type ModelPrice struct {
	InputPerMillion  float64 `yaml:"input_per_million"`
	OutputPerMillion float64 `yaml:"output_per_million"`
	// Cache rates are explicit figures rather than multipliers of the input
	// rate: providers do not agree on a ratio, and an explicit number can be
	// checked against a published price list without arithmetic. Zero means
	// "not declared", which warns rather than defaulting.
	CacheReadPerMillion  float64 `yaml:"cache_read_per_million"`
	CacheWritePerMillion float64 `yaml:"cache_write_per_million"`
}

type config struct {
	Models map[string]ModelPrice `yaml:"models"`
}

type Loader struct {
	models map[string]ModelPrice
}

func NewLoader(overridePath string) (*Loader, error) {
	defaults, err := parseYAML(defaultsYAML)
	if err != nil {
		return nil, fmt.Errorf("parse default pricing: %w", err)
	}

	merged := make(map[string]ModelPrice, len(defaults.Models))
	maps.Copy(merged, defaults.Models)

	if overridePath != "" {
		data, readErr := os.ReadFile(overridePath)
		if readErr != nil && !os.IsNotExist(readErr) {
			return nil, fmt.Errorf("read pricing override: %w", readErr)
		}

		if readErr == nil {
			overrides, parseErr := parseYAML(data)
			if parseErr != nil {
				return nil, fmt.Errorf("parse pricing override: %w", parseErr)
			}

			maps.Copy(merged, overrides.Models)
		}
	}

	return &Loader{models: merged}, nil
}

// CostResult carries what a cost figure is worth, not just its value. A caller
// needs to distinguish "no price for this model", "priced in full" and "priced
// but a tier was skipped", because only the last one produces a number that is
// right about what it covers and wrong about the total.
type CostResult struct {
	Cost  float64
	Known bool
	// MissingTiers names the tiers that carried tokens but had no rate, so the
	// warning can say which rate to add.
	MissingTiers []string
}

// Incomplete reports whether tokens were present in a tier that had no rate and
// were therefore left out of Cost.
func (r CostResult) Incomplete() bool {
	return len(r.MissingTiers) > 0
}

const perMillion = 1_000_000

// ComputeCost prices a usage value tier by tier.
//
// Reasoning is deliberately absent from the computation: it is a subset of
// Output and is already billed inside it.
//
// A tier with tokens but no declared rate is excluded and reported in
// MissingTiers. Charging cache reads at the input rate overstates cost by
// roughly an order of magnitude, and assuming a ratio silently encodes one
// provider's pricing model, so neither is used as a fallback.
func (l *Loader) ComputeCost(model string, u usage.Usage) CostResult {
	price, ok := l.models[model]
	if !ok {
		return CostResult{}
	}

	res := CostResult{Known: true}

	res.Cost = float64(u.InputFresh)/perMillion*price.InputPerMillion +
		float64(u.Output)/perMillion*price.OutputPerMillion

	if u.CacheRead > 0 {
		if price.CacheReadPerMillion > 0 {
			res.Cost += float64(u.CacheRead) / perMillion * price.CacheReadPerMillion
		} else {
			res.MissingTiers = append(res.MissingTiers, "cache_read_per_million")
		}
	}

	if u.CacheWrite > 0 {
		if price.CacheWritePerMillion > 0 {
			res.Cost += float64(u.CacheWrite) / perMillion * price.CacheWritePerMillion
		} else {
			res.MissingTiers = append(res.MissingTiers, "cache_write_per_million")
		}
	}

	if res.Incomplete() {
		slog.Warn(
			"model has cache tokens but no cache rate; those tokens are excluded from cost",
			"model", model,
			"missing", res.MissingTiers,
		)
	}

	return res
}

// CacheSavings is what the cached tokens would have cost had every one been
// billed as fresh input, minus what they actually cost. It is the figure that
// makes prompt caching visible, and it is only computable once tiers exist.
func (l *Loader) CacheSavings(model string, u usage.Usage) (float64, bool) {
	price, ok := l.models[model]
	if !ok {
		return 0, false
	}

	cached := float64(u.CacheRead+u.CacheWrite) / perMillion
	if cached == 0 {
		return 0, true
	}

	asFresh := cached * price.InputPerMillion

	actual := float64(u.CacheRead)/perMillion*price.CacheReadPerMillion +
		float64(u.CacheWrite)/perMillion*price.CacheWritePerMillion

	return asFresh - actual, true
}

func (l *Loader) All() map[string]ModelPrice {
	out := make(map[string]ModelPrice, len(l.models))
	maps.Copy(out, l.models)

	return out
}

func (l *Loader) SetOverride(overridePath, model string, price ModelPrice) error {
	data, err := os.ReadFile(overridePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read pricing override: %w", err)
	}

	var cfg config
	if err == nil {
		if err = yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse pricing override: %w", err)
		}
	}

	if cfg.Models == nil {
		cfg.Models = make(map[string]ModelPrice)
	}

	cfg.Models[model] = price

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal pricing override: %w", err)
	}

	if err = os.WriteFile(overridePath, out, 0o600); err != nil {
		return fmt.Errorf("write pricing override: %w", err)
	}

	l.models[model] = price

	return nil
}

func parseYAML(data []byte) (config, error) {
	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return config{}, err
	}

	return cfg, nil
}
