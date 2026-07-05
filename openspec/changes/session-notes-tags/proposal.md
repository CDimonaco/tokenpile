## Why

When reviewing token usage, raw numbers alone don't explain why an issue cost what it did. Adding a short note and categorical tags to each session lets the developer (and their agent) capture intent alongside the data, turning the usage log into a lightweight work diary.

## What Changes

- `tokenpile log` gains two optional flags: `--note <text>` and `--tag <name>` (repeatable)
- The `sessions` table gains `note TEXT` and `tags TEXT` (JSON array) columns — **BREAKING** schema migration required
- When multiple `log` calls share a session, the last `--note` wins; tags accumulate (union, no duplicates)
- Agent skill templates are updated to pass a one-line `--note` and relevant `--tag` values on each call
- TUI detail view gains a two-tab layout: **Summary** (existing aggregated view) and **Sessions** (per-session list with note and tag badges)
- `tokenpile report` gains an optional `--sessions` flag to print the session list instead of the aggregated table
- Export JSON schema bumped to v2: adds a `sessions` array alongside `entries`; each session carries `note` and `tags`

## Capabilities

### New Capabilities

- `session-annotations`: optional note and tags on sessions, written via `log` flags, accumulated across calls within the same session

### Modified Capabilities

- `sessions`: schema gains `note` and `tags` columns; session domain type updated
- `usage-tracking`: `log` command gains `--note` and `--tag` flags
- `tui`: detail view gains Summary/Sessions tab layout; Sessions tab shows per-session note and tag badges
- `export`: JSON schema v2 adds `sessions` block with note and tags per session
- `agent-skill`: skill templates updated to pass `--note` and `--tag` on log calls

## Impact

- `internal/usage/types.go`: `Session` gains `Note string` and `Tags []string`
- `internal/store/store.go` + `internal/store/sqlite.go`: schema migration, updated insert/query for sessions
- `cmd/tokenpile/cmd_log.go`: `--note` and `--tag` flags; pass through to store
- `internal/tui/tui.go`: tab component for detail view; sessions list renderer
- `cmd/tokenpile/cmd_report.go`: `--sessions` flag
- `internal/export/`: schema v2, sessions block in marshaling
- `internal/skill/templates/`: updated skill instructions
- `internal/mocks/`: regenerated
