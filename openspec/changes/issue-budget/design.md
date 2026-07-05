## Context

tokenpile tracks token cost per issue but provides no way to set expectations upfront. A developer starting work on an issue has no signal that they've spent more than intended until they open the TUI and do the mental arithmetic themselves. A budget is a single number stored against an issue that unlocks visual overage indicators across all output surfaces.

## Goals / Non-Goals

**Goals:**
- `tokenpile budget set` and `budget unset` CLI commands
- Budget stored as a float (USD) keyed by `(repo, issue_num)`
- TUI issue list shows a progress bar with color coding (green / yellow / red)
- TUI detail Summary tab shows budget status and overage amount
- `tokenpile report` includes a budget line when a budget is set
- Export JSON includes an optional `budget` object per issue

**Non-Goals:**
- Budget alerts / notifications (no daemon, no webhooks)
- Team-shared budgets or repo-level budgets
- Budget history or change tracking
- Blocking `log` when over budget

## Decisions

### D1: Separate `issue_budgets` table

Budget is optional and independent from usage data. A separate table with `(repo, issue_num)` as primary key keeps the schema clean and makes `unset` a simple `DELETE`. No nullable column on `usage_entries` or sessions.

### D2: Budget computed at read time, not stored

The "spent" amount is always computed from usage entries — same as `TotalCost` in a report. The budget table stores only the `amount` threshold. This means no sync issue between stored cost and actual cost.

### D3: Progress bar thresholds: 80% yellow, 100% red

Green below 80% (on track), yellow 80–99% (approaching limit), red 100%+ (over). These are common UX conventions for quota indicators and require no configuration.

### D4: Export includes budget as optional top-level field per issue

When exporting for a specific issue (`--issue N`), if a budget is set the export document gains a `budget` object: `{ "amount": 5.00, "spent": 3.20, "over": false }`. When exporting all issues or no budget is set, the field is omitted. The signature is still over `entries` only — budget is context, not audited data.

## Risks / Trade-offs

- [Risk] Budget amount is in USD but pricing config may be incomplete for some models → Mitigation: show `$?? / $5.00` with a note when cost cannot be computed, rather than showing $0
- [Risk] User sets budget then pricing config changes → Mitigation: cost is always recomputed at read time, so budget comparison is always current

## Migration Plan

1. Add `CREATE TABLE IF NOT EXISTS issue_budgets (repo TEXT, issue_num INTEGER, amount REAL, PRIMARY KEY (repo, issue_num))` to schema init — purely additive, no migration needed
2. No existing data affected
