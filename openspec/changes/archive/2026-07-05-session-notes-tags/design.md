## Context

The `sessions` table currently stores only timing data (`started_at`, `ended_at`). Each session groups multiple `log` calls under the same 30-minute idle window. The `log` command writes an `Entry` and may start/reuse a session, but sessions carry no semantic content — no description of what was worked on.

The goal is to add `note` and `tags` to sessions so both the developer and their agent can annotate work context. The challenge is that a session is built up incrementally across multiple `log` calls, so the merge semantics must be explicit.

## Goals / Non-Goals

**Goals:**
- `log` accepts optional `--note` and `--tag` (repeatable) flags
- Each `log` call that belongs to a session can update the session's note and tags
- Last `--note` passed within a session wins; tags are unioned across all calls
- TUI detail view gains a Sessions tab alongside the existing Summary tab
- `report` gains `--sessions` flag for text output
- Export JSON v2 adds a `sessions` array with note and tags per session

**Non-Goals:**
- Per-entry notes (notes live on sessions only)
- Note history / audit trail within a session
- Tag taxonomy enforcement at the CLI level (any string is valid)
- Editing notes/tags after the fact via a dedicated command (can add later)

## Decisions

### D1: Note and tags on Session, not Entry

Entries are API-call granularity. A session is a work block. Notes describe what was being worked on — that's session-level intent, not per-call. Storing on Entry would require joining and de-duplicating on every read.

Alternative: store on Entry and aggregate at read time. Rejected — more complex queries, ambiguous when entries in the same session have different notes.

### D2: Last note wins, tags union

Within a session, the agent calls `log` multiple times. Each call may pass `--note`. Rather than concatenating (produces noise) or first-wins (misses later context), last-wins lets the final call summarize the completed work. Tags accumulate because work can span multiple categories.

Alternative: concatenate notes with newlines. Rejected — produces unreadable multi-line strings in the TUI.

### D3: Tags stored as JSON array in TEXT column

SQLite has no native array type. JSON array in TEXT is simple, readable in sqlite3 CLI, and avoids a separate `session_tags` join table. Tags are always read with the session, never queried independently, so no index needed.

Alternative: separate `session_tags` table. Rejected — overkill for a list that is always fetched alongside the session.

### D4: Schema migration via ALTER TABLE

SQLite supports `ALTER TABLE ADD COLUMN`. Both columns are nullable (no default required for existing rows). No data loss, no full migration needed.

### D5: Export schema v2 — additive, not replacing entries

The existing `entries` array stays unchanged (signature is still over entries). A new `sessions` array is added alongside it. This keeps verification logic intact — the signature covers only entries, as before.

Alternative: include sessions in the signature. Considered but deferred — complicates verification for consumers who only care about entries. Sessions are supplementary context, not the audited data.

### D6: TUI tab navigation with `tab` key

The detail view gains two tabs: Summary (existing) and Sessions (new). `tab` cycles between them. The existing key `?` for help and `q` to quit remain unchanged. Tab state is not persisted across issue navigations.

## Risks / Trade-offs

- [Risk] Agents may pass very long `--note` strings → Mitigation: truncate to 200 chars in the store layer, document limit in skill templates
- [Risk] Tags with spaces or special chars could break shell parsing → Mitigation: document that tags should be single words; no enforcement at CLI level
- [Risk] Export v2 breaks existing consumers parsing the JSON → Mitigation: `schema_version` field already exists; bump to `"2.0"`. Old consumers should check this field.

## Migration Plan

1. `ALTER TABLE sessions ADD COLUMN note TEXT` — safe, nullable, zero downtime
2. `ALTER TABLE sessions ADD COLUMN tags TEXT` — same
3. Existing rows get NULL note and NULL tags — rendered as empty in TUI/report

No rollback needed — columns are additive and nullable. Downgrading the binary leaves the columns populated but ignored.
