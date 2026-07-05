## Why

Without a spending limit, it's easy to lose track of how much an issue is costing until well after the fact. A per-issue budget lets the developer set a threshold upfront and get a clear visual signal — in the TUI, report, and export — when they've gone over.

## What Changes

- New `tokenpile budget` command group with `set` and `unset` subcommands
- `issue_budgets` table added to the SQLite DB: `(repo, issue_num) → amount`
- TUI issue list shows a progress bar (`$0.43 / $5.00`) with color coding: green under 80%, yellow 80–100%, red over 100%
- TUI detail view Summary tab shows budget status and overage when a budget is set
- `tokenpile report` shows budget line when a budget exists for the issue
- Export JSON includes a `budget` object per issue when set: `{ "amount": 5.00, "spent": 0.43, "over": false }`

## Capabilities

### New Capabilities

- `issue-budget`: per-issue spending limit stored in DB, surfaced in TUI, report, and export with color-coded overage indicators

### Modified Capabilities

- `tui`: issue list gains budget progress bar column; detail view Summary tab gains budget status block
- `export`: JSON export includes optional `budget` field per issue

## Impact

- `internal/store/store.go` + `internal/store/sqlite.go`: new `SetBudget`, `UnsetBudget`, `GetBudget` methods; schema migration for `issue_budgets` table
- `cmd/tokenpile/`: new `cmd_budget.go` with `budget set` and `budget unset` subcommands
- `internal/tui/tui.go`: budget progress bar in issue list; budget status in detail Summary tab
- `cmd/tokenpile/cmd_report.go`: budget line in output when budget is set
- `internal/export/`: optional `budget` field in JSON marshaling
- `internal/mocks/`: regenerated
