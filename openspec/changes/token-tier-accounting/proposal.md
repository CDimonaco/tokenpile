## Why

tokenpile models a call as two numbers, `tokens_in` and `tokens_out`, and prices them with two rates. Providers do not bill that way. Anthropic bills cached prompt reads at 10% of the input rate and cache writes at 125%; OpenAI discounts cached prompt tokens and folds reasoning tokens into completions. With two buckets these tiers are not merely rounded, they are unrepresentable.

Measured against a real session on this machine (44.6M tokens, Claude Opus):

| tier | tokens | % tokens | cost | % cost |
|---|---|---|---|---|
| fresh input | 609 | 0.0% | $0.00 | 0.0% |
| cache write | 477,012 | 1.1% | $2.98 | 9.8% |
| cache read | 43,907,365 | 98.4% | $21.95 | 71.8% |
| output | 225,365 | 0.5% | $5.63 | 18.4% |
| **total** | **44,610,351** | | **$30.57** | |

Charging the whole 44.6M at the input rate yields $227.56, a 7.4x overstatement. Today's estimate, produced by asking the model to guess its own context at four characters per token, lands around $0.50 — a 61x understatement. Neither number is a rounding error, and the export signs both with the same Ed25519 key.

Two facts in that table are the reason the feature exists, and neither is expressible today: output is 0.5% of tokens but 18.4% of cost, and caching saved $196.99 on this session, 87% of the un-cached price. With two buckets tokenpile is not losing precision, it is losing the question.

A second gap surfaced alongside: nothing records **how** a count was obtained. An agent cannot fill the tiers, because it cannot see its own cache — the distinction between a measured count and an estimated one is structural, not a transitional defect, so it belongs in the schema.

## What Changes

- **BREAKING** `usage.Entry` carries a `Usage` value type instead of `TokensIn`/`TokensOut`: `InputFresh`, `CacheWrite`, `CacheRead`, `Output`, `Reasoning`. `Reasoning` is a **subset** of `Output` and is never added on top of it.
- **BREAKING** `usage_entries` is recreated with per-tier columns. No incremental migration path is provided: the whole dataset is one test row.
- **BREAKING** Pricing gains two rates per model — `cache_read_per_million` and `cache_write_per_million` — as explicit figures rather than multipliers, so any provider's scheme can be expressed and each number is checkable against a published price list. All 24 models in the defaults are filled in, and `claude-opus-5`, currently missing entirely, is added.
- **BREAKING** `ComputeCost(model string, tokensIn, tokensOut int)` becomes `ComputeCost(model string, u usage.Usage)`. Four tiers make a positional signature untenable.
- A model that has entries with cache tokens but no cache rates SHALL warn rather than silently pick a default, reusing the existing unknown-model warning pattern.
- New `source` field on every entry: `measured` (read from an agent transcript) or `estimated` (declared by a model). It is derived from the command that writes the entry, never from a flag, so it cannot be set wrongly by accident.
- **BREAKING** `tokenpile log` replaces `--tokens-in`/`--tokens-out` with per-tier flags. Entries written by `log` are always `estimated`.
- **BREAKING** Export moves to `schema_version: "4.0"` with per-tier entry fields and `source`. Verification of 2.0 and 3.0 documents is retained: it is already implemented and costs nothing to keep.
- `tokenpile report` shows total cost and cache savings by default; the per-tier breakdown moves behind `--detail`.
- `tokenpile pricing set` gains `--cache-read` and `--cache-write` flags.
- The agent skill template stops instructing agents to estimate token counts.

## Capabilities

### Modified Capabilities

- `usage-tracking`: the `log` command's token flags become per-tier; entries carry a usage tier breakdown and a `source`.
- `pricing`: cost is computed from four tiers with four rates; new requirement for missing cache rates; `pricing set` accepts cache rates.
- `export`: schema 4.0 carries per-tier fields and `source`.

## Impact

- `internal/usage/types.go`: `Usage` introduced; `Entry`, `ReportRow`, `Report`, `TrackedIssue` and `Point` reshaped around it.
- `internal/store/sqlite.go`: `usage_entries` recreated with tier columns plus `source`; roughly six aggregation sites (`ListEntries`, `Report`, `ListIssues`, `OverTime`, and the export listings) change from `SUM(tokens_in)` to per-tier sums.
- `internal/pricing/`: `ModelPrice` gains two fields; `ComputeCost` changes signature; `pricing.defaults.yaml` gains two rates for 24 models and a new `claude-opus-5` entry.
- `internal/export/`: schema 4.0, entry fields, `schema/export.schema.json` updated.
- `cmd/tokenpile/cmd_log.go`, `cmd_report.go`, `cmd_pricing.go`; `internal/tui/` report and chart rendering.
- `internal/skill/templates/`: estimation instructions removed from the remaining agent templates.
- Every existing test asserting `tokens_in`/`tokens_out` is rewritten.
- No data migration: the store is recreated.
