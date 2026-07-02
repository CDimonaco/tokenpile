## ADDED Requirements

### Requirement: Issue list shows assigned and locally-tracked issues

The TUI issue list SHALL display the union of:
1. Issues on GitHub assigned to the authenticated user (`assignee=@me`)
2. Issues in the local DB that have at least one usage entry

The two sources SHALL be merged and deduplicated by `(repo, issue_number)`. Issues present in both sources SHALL appear once. The merged list SHALL be loaded on TUI startup.

#### Scenario: Assigned GitHub issues appear in list
- **WHEN** the user is authenticated
- **WHEN** GitHub returns issues assigned to `@me` for the current repo
- **THEN** those issues appear in the TUI issue list

#### Scenario: Locally-tracked issues appear even if not assigned
- **WHEN** an issue has usage entries in the local DB
- **WHEN** that issue is not assigned to the user on GitHub (or the user is unauthenticated)
- **THEN** the issue still appears in the TUI issue list

#### Scenario: Deduplication when issue is both assigned and tracked locally
- **WHEN** an issue is assigned to the user on GitHub
- **WHEN** the same issue has usage entries in the local DB
- **THEN** the issue appears exactly once in the list

### Requirement: Unauthenticated fallback to DB-only list

When no valid auth token exists, the TUI SHALL skip the GitHub API call and populate the issue list solely from local DB usage entries. The TUI SHALL display a status line indicating the user is not authenticated and showing local issues only. The TUI SHALL NOT show an error screen or block the user from browsing local data.

#### Scenario: DB-only list when unauthenticated
- **WHEN** no valid OAuth token exists
- **THEN** the issue list shows only issues with local usage entries
- **THEN** a status line reads "(not authenticated — showing local issues only)"
- **THEN** the TUI does not display an auth error screen

#### Scenario: Empty state when unauthenticated and no local data
- **WHEN** no valid OAuth token exists
- **WHEN** no usage has been logged
- **THEN** the list view shows the empty state message

### Requirement: DB-only issue stub when GitHub issue is inaccessible

If an issue exists in the local DB but cannot be fetched from GitHub (deleted, transferred, private repo after access revoked), the TUI SHALL display a stub row using the stored repo and issue number. The stub SHALL show `[not found on GitHub]` in place of the title.

#### Scenario: Stub row for inaccessible GitHub issue
- **WHEN** a DB-tracked issue returns 404 or access denied from GitHub
- **THEN** the issue appears in the list as `#N [not found on GitHub]`
- **THEN** the TUI does not crash or hide the issue
