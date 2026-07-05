## 1. Store

- [ ] 1.1 Add `CREATE TABLE IF NOT EXISTS issue_budgets (repo TEXT NOT NULL, issue_num INTEGER NOT NULL, amount REAL NOT NULL, PRIMARY KEY (repo, issue_num))` to the `schema` const in `internal/store/sqlite.go` (picked up by `Open` on first run; existing DBs get the table via `CREATE TABLE IF NOT EXISTS` which is already idempotent — no ALTER needed here)
- [ ] 1.2 Add `SetBudget(ctx context.Context, repo string, issueNum int, amount float64) error` to `store.Store` interface
- [ ] 1.3 Add `UnsetBudget(ctx context.Context, repo string, issueNum int) error` to `store.Store` interface
- [ ] 1.4 Add `GetBudget(ctx context.Context, repo string, issueNum int) (float64, bool, error)` to `store.Store` interface (bool = exists)
- [ ] 1.5 Implement `SetBudget` in `internal/store/sqlite.go` using `INSERT OR REPLACE`
- [ ] 1.6 Implement `UnsetBudget` in `internal/store/sqlite.go` using `DELETE`; no error if row does not exist
- [ ] 1.7 Implement `GetBudget` in `internal/store/sqlite.go`

## 2. Store: tests

- [ ] 2.1 Unit test `SetBudget`: set, then overwrite with new amount
- [ ] 2.2 Unit test `UnsetBudget`: unset existing; unset non-existing (idempotent)
- [ ] 2.3 Unit test `GetBudget`: exists returns correct amount; not exists returns false

## 3. Mocks

- [ ] 3.1 Run `make generate` to regenerate mocks for updated `Store` interface

## 4. Budget command

- [ ] 4.1 Create `cmd/tokenpile/cmd_budget.go` with a `budgetCommand` function returning a `*cli.Command` named `budget`
- [ ] 4.2 Add `budget set` subcommand: flags `--issue` (required), `--repo` (optional, inferred), `--amount` (required, float > 0); calls `store.SetBudget`; prints confirmation
- [ ] 4.3 Add `budget unset` subcommand: flags `--issue` (required), `--repo` (optional, inferred); calls `store.UnsetBudget`; prints confirmation
- [ ] 4.4 Register `budgetCommand` in `cmd/tokenpile/main.go`

## 5. Report command

- [ ] 5.1 In `cmd/tokenpile/cmd_report.go`, after printing totals, call `store.GetBudget`; if a budget is set print a budget line: `Budget: $X.XX / $Y.YY (N%)` with overage if over

## 6. TUI: issue list

- [ ] 6.1 Update `TrackedIssue` loading in TUI to also fetch budget via `store.GetBudget` for each issue (or batch-load with a new `ListBudgets` store method if N issues is large — use per-issue fetch for now)
- [ ] 6.2 Add budget indicator column to the issue list renderer: `$spent / $budget` colored green/yellow/red per thresholds; empty string when no budget
- [ ] 6.3 Update the list header row to include the budget column

## 7. TUI: detail view

- [ ] 7.1 On detail view load, fetch budget via `store.GetBudget` and store in model
- [ ] 7.2 In the Summary tab renderer, if budget is set render a budget status line below totals: `Budget: $spent / $amount (N%)` — red and with overage text when over

## 8. TUI: tests

- [ ] 8.1 Unit test issue list render: issue with budget shows colored indicator; issue without budget shows no indicator
- [ ] 8.2 Unit test detail Summary tab: budget line present when set; absent when not set; overage text when over

## 9. Export

- [ ] 9.1 Add optional `budget` struct to `internal/export/export.go`: `BudgetAmount float64`, `BudgetSpent float64`, `BudgetOver bool`
- [ ] 9.2 Update `Build` to accept an optional `*BudgetInfo` parameter; when non-nil and export is for a single issue, include `budget` in the document (outside the signed entries)
- [ ] 9.3 Update the caller in `cmd/tokenpile/cmd_export.go` to pass budget info when `--issue` flag is set

## 10. Export: tests

- [ ] 10.1 Unit test `Build` with budget: document contains `budget` field; signature unchanged vs same build without budget
- [ ] 10.2 Unit test `Build` without budget: document omits `budget` field
