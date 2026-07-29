## 1. Usage type

- [x] 1.1 Add `Usage{InputFresh, CacheWrite, CacheRead, Output, Reasoning}` to `internal/usage/types.go` with a doc comment stating that `Reasoning` is a subset of `Output` and is never added to it
- [x] 1.2 Replace `TokensIn`/`TokensOut` with a `Usage` field on `Entry`, `ReportRow` and `Point`; replace `TotalTokensIn`/`TotalTokensOut` on `Report` and `TrackedIssue` with a `Usage` total
- [x] 1.3 Add a `Source` type with `SourceMeasured`/`SourceEstimated` and a `Source` field on `Entry`
- [x] 1.4 Add helper for the total token count used by report displays, computed from tiers, excluding `Reasoning`

## 2. Store

- [x] 2.1 Rewrite the `usage_entries` `CREATE TABLE`: `input_fresh`, `cache_write`, `cache_read`, `output`, `reasoning`, `source`, all `NOT NULL`; drop `tokens_in`/`tokens_out`. Leave the `migrations` slice in place for future additive changes
- [x] 2.2 Update `AddEntry` insert and every row scan
- [x] 2.3 Update all aggregation sites from `SUM(tokens_in)`/`SUM(tokens_out)` to per-tier sums: `Report`, `ListIssues`, `OverTime`, entry listings, and the export listings
- [x] 2.4 Run `make generate` to regenerate store mocks
- [x] 2.5 Store tests: tiers round-trip independently, source round-trips, aggregations sum per tier

## 3. Pricing

- [x] 3.1 Add `CacheReadPerMillion` and `CacheWritePerMillion` to `ModelPrice`
- [x] 3.2 Change `ComputeCost` to `(model string, u usage.Usage) (float64, bool)`; bill four tiers and exclude `Reasoning`; update every caller
- [x] 3.3 Return enough information for callers to know a rate was missing (not just a bool), so the warning can name the model and the tier
- [x] 3.4 Emit the missing-cache-rate WARN and mark the figure incomplete in report output
- [x] 3.5 Add `CacheSavings(model, u)`: what the cached tokens would have cost at the fresh input rate, minus what they cost
- [x] 3.6 Fill `cache_read_per_million` and `cache_write_per_million` for all 24 models in `pricing.defaults.yaml`, checking each against the provider's published price list
- [x] 3.7 Add the missing `claude-opus-5` entry to the defaults
- [x] 3.8 `pricing set`: add `--cache-read` and `--cache-write`; `pricing list` shows four rates
- [x] 3.9 Pricing tests: four-tier cost matches the worked example ($30.57), reasoning does not change cost, missing rate warns and excludes the tier, savings computed correctly

## 4. Export

- [x] 4.1 Bump `SchemaVersion` to `4.0`; entry JSON carries the five tier fields and `source`, and no `tokens_in`/`tokens_out`
- [x] 4.2 Remove the legacy verification path entirely: drop `legacySchemaVersion`, the entries-only digest branch, `VerifyResult.Legacy` and its handling in `cmd_export.go`; unknown versions fail with an error naming the version found
- [x] 4.3 Update `schema/export.schema.json`
- [x] 4.4 Export tests: 4.0 round-trips and verifies, tampering with a tier invalidates the signature, a pre-4.0 document is rejected with a clear error; delete the `testdata/export_v2.json` fixture and its tests

## 5. CLI and TUI

- [x] 5.1 `cmd_log.go`: replace `--tokens-in`/`--tokens-out` with `--input`, `--cache-write`, `--cache-read`, `--output`, `--reasoning`; require at least one; write `SourceEstimated`; reject any attempt to set the source
- [x] 5.2 `cmd_report.go`: default output shows total cost and cache savings; `--detail` shows the per-tier table with token share and cost share
- [x] 5.3 `internal/tui/`: update report and chart rendering for the new totals
- [x] 5.4 CLI tests: log with each tier flag, log with no token flag fails, report default and `--detail` shapes

## 6. Skill templates and docs

- [x] 6.1 Remove the token estimation instructions from the remaining agent skill templates; bump the skill version marker
- [x] 6.2 README: new `log` flags, report output, pricing config shape with cache rates
- [x] 6.3 CLAUDE.md: key design decisions gain the tier model, reasoning-as-subset, and provenance
- [x] 6.4 Run `make check` and commit
