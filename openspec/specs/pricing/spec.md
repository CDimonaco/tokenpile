
### Requirement: Embedded default pricing configuration

The system SHALL embed a default pricing YAML file (`pricing.defaults.yaml`) in the binary using Go's `embed` package. The defaults file SHALL include pricing for commonly used models at time of implementation, covering at minimum: Claude models (Anthropic) and GPT models (OpenAI). Prices SHALL be expressed in USD per 1,000,000 tokens, separately for input and output.

#### Scenario: Default pricing available without user configuration
- **WHEN** no user pricing override file exists
- **WHEN** cost is computed for a known model
- **THEN** the embedded default price is used

### Requirement: User pricing override file

The system SHALL support a user-managed pricing override file at `~/.config/tokenpile/pricing.yaml`. This file uses the same YAML structure as the embedded defaults. At runtime, the system SHALL merge the two: user entries take precedence over defaults, and defaults fill in models not present in the user file.

Structure:
```yaml
models:
  claude-sonnet-4-6:
    input_per_million: 3.00
    output_per_million: 15.00
  gpt-4o:
    input_per_million: 2.50
    output_per_million: 10.00
```

#### Scenario: User override takes precedence
- **WHEN** the user file defines a price for `claude-sonnet-4-6`
- **THEN** the user-defined price is used instead of the embedded default

#### Scenario: Default fills missing model
- **WHEN** the user file does not define a price for `gpt-4o`
- **THEN** the embedded default price for `gpt-4o` is used

#### Scenario: Override file does not exist
- **WHEN** `~/.config/tokenpile/pricing.yaml` does not exist
- **THEN** only the embedded defaults are used and no error is returned

### Requirement: Cost computation at report time

The system SHALL compute cost at the time a report or export is generated, not at log time. Cost SHALL NOT be stored in `usage_entries`. For a given `UsageEntry`, cost SHALL be computed as:

```
cost = (tokens_in / 1_000_000 * input_per_million)
     + (tokens_out / 1_000_000 * output_per_million)
```

#### Scenario: Cost computed correctly
- **WHEN** an entry has 10,000 tokens in and 2,000 tokens out
- **WHEN** the model price is $3.00/1M in and $15.00/1M out
- **THEN** the computed cost is $0.030 + $0.030 = $0.060

### Requirement: Unknown model warning

The system SHALL warn (via `slog` at WARN level and in report output) when a `UsageEntry` references a model not present in either the embedded defaults or the user override file. Cost for unknown models SHALL be reported as $0.00 with a visual indicator.

#### Scenario: Unknown model in report
- **WHEN** an entry references model `my-local-llm`
- **WHEN** no pricing entry exists for `my-local-llm`
- **THEN** the report shows cost as $0.00 with a note: "pricing unknown"
- **THEN** a WARN log entry is emitted

### Requirement: Pricing management commands

The system SHALL provide `tokenpile pricing list` to display the merged pricing config (defaults + overrides) in a human-readable table.

The system SHALL provide `tokenpile pricing set <model> --in <price> --out <price>` to add or update a model's price in the user override file.

#### Scenario: List shows merged config
- **WHEN** `tokenpile pricing list` is run
- **THEN** all models from both files are shown with their effective prices
- **THEN** overridden models are visually distinguished from defaults

#### Scenario: Set adds new model price
- **WHEN** `tokenpile pricing set my-local-llm --in 0.00 --out 0.00`
- **THEN** `my-local-llm` is added to the user override file
- **THEN** subsequent reports use $0.00 for that model without warnings
