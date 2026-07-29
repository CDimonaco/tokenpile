## ADDED Requirements

### Requirement: Missing cache rates warn rather than default

When an entry carries cache read or cache write tokens for a model whose pricing declares no corresponding rate, the system SHALL warn via `slog` at WARN level and mark the affected figure in report output, and SHALL exclude that tier from the computed cost rather than substituting the input rate or an assumed ratio.

Treating cache reads at the full input rate overstates cost by roughly an order of magnitude; assuming a ratio silently encodes one provider's pricing model. An incomplete figure that says so is preferable to a confident wrong one.

#### Scenario: Cache tokens without cache rates
- **WHEN** an entry has cache read tokens for a model with `input_per_million` and `output_per_million` but no `cache_read_per_million`
- **THEN** a WARN log entry names the model and the missing rate
- **THEN** the report marks the cost figure as incomplete
- **THEN** the cache read tokens contribute nothing to the computed cost

#### Scenario: Complete rates produce no warning
- **WHEN** every tier present in an entry has a corresponding rate
- **THEN** no warning is emitted

### Requirement: Cache savings

The system SHALL be able to report, for any scope it reports cost over, the amount saved by prompt caching: the cost the same tokens would have incurred had every cached token been billed at the fresh input rate, minus the cost actually computed.

#### Scenario: Savings computed from tiers
- **WHEN** a scope contains 43,907,365 cache read tokens for a model priced at $5.00/1M input and $0.50/1M cache read
- **THEN** the reported saving for those tokens is $197.58

#### Scenario: No caching means no saving
- **WHEN** a scope contains no cache read or cache write tokens
- **THEN** the reported saving is zero

## MODIFIED Requirements

### Requirement: Cost computation at report time

The system SHALL compute cost at the time a report or export is generated, not at log time. Cost SHALL NOT be stored in `usage_entries`. For a given entry, cost SHALL be computed from the usage tiers as:

```
cost = (input_fresh  / 1_000_000 * input_per_million)
     + (cache_write  / 1_000_000 * cache_write_per_million)
     + (cache_read   / 1_000_000 * cache_read_per_million)
     + (output       / 1_000_000 * output_per_million)
```

Reasoning tokens SHALL NOT appear in this computation: they are a subset of output and are already billed as part of it.

`ComputeCost` SHALL take the model name and a usage value rather than positional token counts.

#### Scenario: Cost computed correctly
- **WHEN** an entry has 10,000 fresh input tokens and 2,000 output tokens
- **WHEN** the model price is $3.00/1M in and $15.00/1M out
- **THEN** the computed cost is $0.030 + $0.030 = $0.060

#### Scenario: Cost across all four tiers
- **WHEN** an entry has 609 fresh input, 477,012 cache write, 43,907,365 cache read and 225,365 output tokens
- **WHEN** the model is priced at $5.00/1M input, $6.25/1M cache write, $0.50/1M cache read and $25.00/1M output
- **THEN** the computed cost is $30.57

#### Scenario: Reasoning tokens are not billed twice
- **WHEN** an entry has 1,000 output tokens of which 400 are reasoning tokens
- **THEN** the cost counts 1,000 output tokens, not 1,400

### Requirement: Embedded default pricing configuration

The system SHALL embed a default pricing configuration covering the models it knows about. Each model SHALL declare `input_per_million`, `output_per_million`, `cache_read_per_million` and `cache_write_per_million` as explicit figures, so each can be checked against a provider's published price list without arithmetic.

#### Scenario: Default pricing available without user configuration
- **WHEN** tokenpile runs with no user pricing override present
- **THEN** cost is computed from the embedded defaults

#### Scenario: Every default model declares four rates
- **WHEN** the embedded defaults are loaded
- **THEN** every model entry declares all four rates

### Requirement: Pricing management commands

The system SHALL provide `tokenpile pricing list` showing the merged configuration and `tokenpile pricing set <model>` writing to the user override file. `set` SHALL accept `--in`, `--out`, `--cache-read` and `--cache-write`.

#### Scenario: List shows merged config
- **WHEN** `tokenpile pricing list` is called
- **THEN** the output shows every known model with all four rates, indicating which come from the override file

#### Scenario: Set adds new model price
- **WHEN** `tokenpile pricing set my-model --in 1.50 --out 6.00 --cache-read 0.15 --cache-write 1.88` is called
- **THEN** the override file contains the model with all four rates
- **THEN** subsequent cost computation uses them

#### Scenario: Set without cache rates
- **WHEN** `tokenpile pricing set my-model --in 1.50 --out 6.00` is called
- **THEN** the model is written with input and output rates only
- **THEN** entries with cache tokens for that model warn per the missing-rates requirement
