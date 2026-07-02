## Why

The TUI issue list currently fetches all issues from a repo without any filter, showing unrelated issues the user is not working on. Users should see only issues relevant to them: those assigned to them on GitHub, plus any issue they have already logged usage against locally.

## What Changes

- `ListIssues` on `GitHubIssueProvider` gains an `assignee` filter defaulting to `@me`
- A new `Store.ListTrackedIssueRefs` method returns `(repo, issue_number)` pairs that have usage entries in the local DB
- The TUI loads both lists and merges them (union, deduplicated by repo+number)
- When unauthenticated, the TUI falls back to DB-only list (no GitHub call)
- Issues from the DB that no longer exist on GitHub (deleted, transferred) show with a `[local]` label instead of failing

## Capabilities

### New Capabilities

- `tui-issue-filter`: TUI issue list filtered to assigned + locally-tracked issues, with unauthenticated fallback to DB-only

### Modified Capabilities

- none

## Impact

- `internal/provider/provider.go`: `IssueProvider.ListIssues` filter struct gains `Assignee` field
- `internal/provider/github_issues.go`: passes `Assignee` to GitHub API options
- `internal/store/store.go`: new `ListTrackedIssueRefs` method on `Store` interface
- `internal/store/sqlite.go`: implements `ListTrackedIssueRefs`
- `internal/tui/tui.go`: merge logic for assigned + DB issues; unauthenticated fallback
- `internal/mocks/`: regenerated mocks for updated interfaces
