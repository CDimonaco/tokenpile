## 1. Usage type

- [ ] 1.1 Add `Usage{InputFresh, CacheWrite, CacheRead, Output, Reasoning}` to `internal/usage/types.go` with a doc comment stating that `Reasoning` is a subset of `Output` and is never added to it
- [ ] 1.2 Replace `TokensIn`/`TokensOut` with a `Usage` field on `Entry`, `ReportRow` and `Point`; replace `TotalTokensIn`/`TotalTokensOut` on `Report` and `TrackedIssue` with a `Usage` total
- [ ] 1.3 Add a `Source` type with `SourceMeasured`/`SourceEstimated` and a `Source` field on `Entry`
- [ ] 1.4 Add helper for the total token count used by report displays, computed from tiers, excluding `Reasoning`

## 2. Store

- [ ] 2.1 Rewrite the `usage_entries` `CREATE TABLE`: `input_fresh`, `cache_write`, `cache_read`, `output`, `reasoning`, `source`, all `NOT NULL`; drop `tokens_in`/`tokens_out`. Leave the `migrations` slice in place for future additive changes
- [ ] 2.2 Update `AddEntry` insert and every row scan
- [ ] 2.3 Update all aggregation sites from `SUM(tokens_in)`/`SUM(tokens_out)` to per-tier sums: `Report`, `ListIssues`, `OverTime`, entry listings, and the export listings
- [ ] 2.4 Run `make generate` to regenerate store mocks
- [ ] 2.5 Store tests: tiers round-trip independently, source round-trips, aggregations sum per tier

## 3. Pricing

- [ ] 3.1 Add `CacheReadPerMillion` and `CacheWritePerMillion` to `ModelPrice`
- [ ] 3.2 Change `ComputeCost` to `(model string, u usage.Usage) (float64, bool)`; bill four tiers and exclude `Reasoning`; update every caller
- [ ] 3.3 Return enough information for callers to know a rate was missing (not just a bool), so the warning can name the model and the tier
- [ ] 3.4 Emit the missing-cache-rate WARN and mark the figure incomplete in report output
- [ ] 3.5 Add `CacheSavings(model, u)`: what the cached tokens would have cost at the fresh input rate, minus what they cost
- [ ] 3.6 Fill `cache_read_per_million` and `cache_write_per_million` for all 24 models in `pricing.defaults.yaml`, checking each against the provider's published price list
- [ ] 3.7 Add the missing `claude-opus-5` entry to the defaults
- [ ] 3.8 `pricing set`: add `--cache-read` and `--cache-write`; `pricing list` shows four rates
- [ ] 3.9 Pricing tests: four-tier cost matches the worked example ($30.57), reasoning does not change cost, missing rate warns and excludes the tier, savings computed correctly

## 4. Export

- [ ] 4.1 Bump `SchemaVersion` to `4.0`; entry JSON carries the five tier fields and `source`, and no `tokens_in`/`tokens_out`
- [ ] 4.2 Keep 2.0 and 3.0 verification paths; verification output names the schema version applied
- [ ] 4.3 Update `schema/export.schema.json`
- [ ] 4.4 Export tests: 4.0 round-trips and verifies, tampering with a tier invalidates the signature, a 3.0 fixture still verifies

## 5. CLI and TUI

- [ ] 5.1 `cmd_log.go`: replace `--tokens-in`/`--tokens-out` with `--input`, `--cache-write`, `--cache-read`, `--output`, `--reasoning`; require at least one; write `SourceEstimated`; reject any attempt to set the source
- [ ] 5.2 `cmd_report.go`: default output shows total cost and cache savings; `--detail` shows the per-tier table with token share and cost share
- [ ] 5.3 `internal/tui/`: update report and chart rendering for the new totals
- [ ] 5.4 CLI tests: log with each tier flag, log with no token flag fails, report default and `--detail` shapes

## 6. Skill templates and docs

- [ ] 6.1 Remove the token estimation instructions from the remaining agent skill templates; bump the skill version marker
- [ ] 6.2 README: new `log` flags, report output, pricing config shape with cache rates
- [ ] 6.3 CLAUDE.md: key design decisions gain the tier model, reasoning-as-subset, and provenance
- [ ] 6.4 Run `make check` and commit
