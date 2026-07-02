## Context

The TUI currently calls `ListIssues` with an empty filter, returning all open issues in the repo. This is noisy and not useful — users want to see issues relevant to their work. Two sources of relevance exist: GitHub (assigned to me) and the local DB (issues I have already logged usage against).

Current flow: `tui.go` → `issueProvider.ListIssues(filter)` → GitHub API → all open issues.

## Goals / Non-Goals

**Goals:**
- Show issues assigned to the authenticated user (`assignee=@me` on GitHub API)
- Always show issues with local usage entries, even if no longer assigned
- Merge the two sources, deduplicate by `(repo, issue_number)`
- Degrade gracefully when unauthenticated: show DB-only list, no error screen

**Non-Goals:**
- Filtering by label, milestone, or other GitHub criteria
- Paginating beyond the first page of assigned issues (can revisit if needed)
- Fetching issue metadata (title, state) for DB-only issues that do not exist on GitHub

## Decisions

### 1. `assignee=@me` via filter struct, not hardcoded in TUI

`usage.Filter` gains an `Assignee string` field. The TUI sets it to `"@me"`. The GitHub provider passes it through to the API. This keeps the provider generic and testable.

Alternatives considered:
- Hardcode `@me` in `github_issues.go`: couples the provider to a single use case, harder to test.
- Add a separate `ListMyIssues` method: interface bloat for what is just a filter parameter.

### 2. New `Store.ListTrackedIssueRefs` for DB-sourced issues

Returns `[]usage.TrackedIssueRef{Repo, IssueNum}` — the minimal data needed for the union. The TUI then calls `GetIssue` on the provider to enrich with title/state, falling back to a stub entry if the issue is no longer accessible.

Alternatives considered:
- Reuse existing `Store.ListIssues`: that method returns aggregated report data; it is heavier and semantically different. A lightweight ref query is cleaner.

### 3. Merge in TUI, not in a new service layer

The TUI is already the composition point for store + provider. Adding a merge helper in `tui.go` keeps the dependency graph flat. No new package needed.

### 4. Unauthenticated fallback: DB-only, no error

If `authToken == ""`, skip the GitHub call entirely and populate the list from DB refs only. Show a dim status line `(not authenticated — showing local issues only)` rather than an error screen.

## Risks / Trade-offs

- **GitHub rate limits**: `@me` assignee query counts against the authenticated user's rate limit (5000 req/h). One call per TUI open is negligible.
- **DB issue with deleted GitHub issue**: `GetIssue` returns 404 → show stub `[#N — not found on GitHub]`. Acceptable edge case.
- **`@me` requires authentication**: If the token is expired or revoked mid-session, the GitHub call fails silently and the TUI falls back to DB-only for that load. A re-login prompt is out of scope here.
