## 1. Domain types

- [ ] 1.1 Add `Note string` and `Tags []string` fields to `usage.Session` in `internal/usage/types.go`

## 2. Store: schema migration

- [ ] 2.1 After the `CREATE TABLE IF NOT EXISTS` block in `internal/store/sqlite.go`, add a `runMigrations` function that executes each `ALTER TABLE` statement and ignores the "duplicate column name" error SQLite returns when the column already exists
- [ ] 2.2 Add migration: `ALTER TABLE sessions ADD COLUMN note TEXT`
- [ ] 2.3 Add migration: `ALTER TABLE sessions ADD COLUMN tags TEXT`
- [ ] 2.4 Call `runMigrations` in `Open` after `schema` is applied, so existing DBs are upgraded on first run of the new binary

## 3. Store: session annotation update

- [ ] 3.1 Add `UpdateSessionAnnotations(ctx context.Context, sessionID string, note *string, tags []string) error` to `store.Store` interface
- [ ] 3.2 Implement `UpdateSessionAnnotations` in `internal/store/sqlite.go`: if `note != nil` update the note column; merge incoming tags with existing tags (union, no duplicates); store tags as JSON array
- [ ] 3.3 Update `ListSessions` in `internal/store/sqlite.go` to scan `note` and `tags` columns into `usage.Session`

## 4. Store: tests

- [ ] 4.1 Unit test `UpdateSessionAnnotations` in `internal/store/sqlite_test.go`: note set, note replaced, tags accumulated, duplicate tags dropped, nil note leaves note unchanged
- [ ] 4.2 Unit test `ListSessions` returns `Note` and `Tags` populated correctly
- [ ] 4.3 Unit test migration idempotency: open a DB, run `Open` a second time, verify no error is returned (columns already exist)

## 5. Mocks

- [ ] 5.1 Run `make generate` to regenerate mocks for updated `Store` interface

## 6. Log command

- [ ] 6.1 Add optional `--note` string flag to `cmd/tokenpile/cmd_log.go`
- [ ] 6.2 Add optional `--tag` string slice flag (repeatable) to `cmd/tokenpile/cmd_log.go`
- [ ] 6.3 After logging the entry, if `--note` or `--tag` flags are present, call `store.UpdateSessionAnnotations` with the session ID from the log response; truncate note to 200 chars before passing

## 7. Skill templates and version detection

- [ ] 7.1 Update `internal/skill/templates/claude-code.md` to include `--note` and `--tag` in the example `tokenpile log` invocation and document the tag vocabulary
- [ ] 7.2 Update `internal/skill/templates/codex.md` with the same additions
- [ ] 7.3 Update `internal/skill/templates/opencode.md` with the same additions
- [ ] 7.4 Add a version header comment to each skill template (e.g. `<!-- tokenpile-skill-version: 2 -->`) so the installed file can be compared against the embedded template
- [ ] 7.5 In `internal/skill/skill.go`, add an `IsUpToDate(agentName string) bool` function that reads the installed file, extracts the version comment, and compares it to the embedded template version
- [ ] 7.6 In `tokenpile skill list`, add an `Up to date` column; show `outdated` (with a hint to re-run `tokenpile skill install --agent <name>`) when `IsUpToDate` returns false

## 8. Report command

- [ ] 8.1 Add optional `--sessions` boolean flag to `cmd/tokenpile/cmd_report.go`
- [ ] 8.2 When `--sessions` is set, fetch sessions via `store.ListSessions` and print a per-session list: start time, end time or "active", duration, tags, note, token totals and cost for that session (computed from entries filtered by session ID)

## 9. TUI: detail view tabs

- [ ] 9.1 Add a `detailTab` int field to `tui.Model` (0 = Summary, 1 = Sessions)
- [ ] 9.2 Handle `tab` key in the detail view to cycle `detailTab` between 0 and 1
- [ ] 9.3 Render tab headers ("Summary" / "Sessions") at the top of the detail view with the active tab visually highlighted
- [ ] 9.4 When `detailTab == 0`, render the existing Summary content unchanged
- [ ] 9.5 When `detailTab == 1`, fetch sessions via `store.ListSessions` (load on first switch, cache in model); render each session as: start→end, duration, tag badges, note (second line if non-empty), token and cost totals
- [ ] 9.6 Add scroll offset field for sessions list; handle up/down arrow keys when Sessions tab is active
- [ ] 9.7 Reset `detailTab` to 0 and sessions cache when navigating to a different issue

## 10. TUI: tests

- [ ] 10.1 Unit test tab switch in detail view: pressing `tab` changes active tab
- [ ] 10.2 Unit test Sessions tab render: session with note and tags shows both; session without annotations shows cleanly

## 11. Export: schema v2

- [ ] 11.1 Add `sessionJSON` struct to `internal/export/export.go` with fields: `ID`, `Repo`, `IssueNum`, `StartedAt`, `EndedAt` (omitempty), `Note` (omitempty), `Tags` (omitempty)
- [ ] 11.2 Update `Document` struct: bump `SchemaVersion` constant to `"2.0"`; add `Sessions []sessionJSON` field
- [ ] 11.3 Update `Build` function signature to accept `[]usage.Session` alongside `[]usage.Entry`; populate `Sessions` in the document; keep signature computation over `entries` only (unchanged)
- [ ] 11.4 Update `internal/schema/` embedded JSON schema to v2: add `sessions` array definition
- [ ] 11.5 Update all callers of `export.Build` to pass sessions (fetch via `store.ListSessions` scoped to the export filter)

## 12. Export: tests

- [ ] 12.1 Unit test `Build` with sessions: document contains sessions array; signature is identical to a build with same entries but different session notes
- [ ] 12.2 Unit test JSON schema v2 validates a document with sessions; validates a document without sessions (sessions omitted or empty)
