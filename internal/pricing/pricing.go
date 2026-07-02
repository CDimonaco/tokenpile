package pricing

import (
	_ "embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed pricing.defaults.yaml
var defaultsYAML []byte

type ModelPrice struct {
	InputPerMillion  float64 `yaml:"input_per_million"`
	OutputPerMillion float64 `yaml:"output_per_million"`
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
	for k, v := range defaults.Models {
		merged[k] = v
	}

	if overridePath != "" {
		data, err := os.ReadFile(overridePath)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read pricing override: %w", err)
		}

		if err == nil {
			overrides, err := parseYAML(data)
			if err != nil {
				return nil, fmt.Errorf("parse pricing override: %w", err)
			}

			for k, v := range overrides.Models {
				merged[k] = v
			}
		}
	}

	return &Loader{models: merged}, nil
}

func (l *Loader) ComputeCost(model string, tokensIn, tokensOut int) (float64, bool) {
	price, ok := l.models[model]
	if !ok {
		return 0, false
	}

	cost := float64(tokensIn)/1_000_000*price.InputPerMillion +
		float64(tokensOut)/1_000_000*price.OutputPerMillion

	return cost, true
}

func (l *Loader) All() map[string]ModelPrice {
	out := make(map[string]ModelPrice, len(l.models))
	for k, v := range l.models {
		out[k] = v
	}

	return out
}

func (l *Loader) SetOverride(overridePath, model string, inputPerM, outputPerM float64) error {
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

	cfg.Models[model] = ModelPrice{
		InputPerMillion:  inputPerM,
		OutputPerMillion: outputPerM,
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal pricing override: %w", err)
	}

	if err = os.WriteFile(overridePath, out, 0o600); err != nil {
		return fmt.Errorf("write pricing override: %w", err)
	}

	l.models[model] = ModelPrice{InputPerMillion: inputPerM, OutputPerMillion: outputPerM}

	return nil
}

func parseYAML(data []byte) (config, error) {
	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return config{}, err
	}

	return cfg, nil
}
