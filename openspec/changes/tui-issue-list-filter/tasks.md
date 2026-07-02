## 1. Store

- [x] 1.1 Add `TrackedIssueRef` type to `internal/usage/types.go` with fields `Repo string` and `IssueNum int`
- [x] 1.2 Add `ListTrackedIssueRefs(ctx context.Context) ([]usage.TrackedIssueRef, error)` to `store.Store` interface
- [x] 1.3 Implement `ListTrackedIssueRefs` in `internal/store/sqlite.go` with a `SELECT DISTINCT repo, issue_num FROM usage_entries` query
- [x] 1.4 Regenerate mocks with `make generate`

## 2. Provider

- [x] 2.1 Add `Assignee string` field to `usage.Filter` in `internal/usage/types.go`
- [x] 2.2 Pass `filter.Assignee` to `github.IssueListByRepoOptions.Assignee` in `internal/provider/github_issues.go`

## 3. TUI

- [x] 3.1 In `tui.go`, after loading, call `store.ListTrackedIssueRefs` to get DB refs
- [x] 3.2 Call `issueProvider.ListIssues` with `Assignee: "@me"` when authenticated; skip when `authToken == ""`
- [x] 3.3 Implement merge/dedup logic: union of GitHub issues and DB refs, keyed by `(repo, issueNum)`
- [x] 3.4 For DB refs not returned by GitHub, call `GetIssue`; on error (404/unauth) use stub `Issue{Number: n, Repo: repo, Title: "[not found on GitHub]"}`
- [x] 3.5 When unauthenticated, show status line `(not authenticated — showing local issues only)` instead of the auth error screen

## 4. Tests

- [x] 4.1 Unit test `ListTrackedIssueRefs` in `internal/store/sqlite_test.go`
- [x] 4.2 Unit test TUI merge logic: assigned + DB, dedup, unauthenticated fallback, stub for inaccessible issue
- [x] 4.3 Update existing TUI auth-error test: unauthenticated now shows DB list + status line, not error screen
