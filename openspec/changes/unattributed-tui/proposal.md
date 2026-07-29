## Why

Capture now records every turn whether or not the issue is known, so unattributed usage accumulates by design. `deterministic-capture` shipped the CLI for it — `tokenpile unattributed` lists groups by session, `assign` and `unassign` move a whole session — and that is enough to never lose a measurement, but not enough to make reconciliation something a person will actually do.

Two gaps remain, and they are related.

**Reconciling is a review task, and the CLI is the wrong shape for it.** Deciding which issue a stretch of work belonged to means looking at when it happened, how much it cost, and what else was going on. That is a list you scan and act on, which is what the TUI already is for every other view in tokenpile. Copying a session UUID from one command into another is the kind of friction that leaves usage unattributed forever.

**Suggestions were specified and could not be delivered.** The original design called for offering the issue a branch name implies. Entries do not record the branch they were captured on, so no suggestion can be derived at listing time. The capture path already knows the branch — Claude Code puts `gitBranch` in every transcript record, and the opencode plugin resolves it — it is simply dropped when a turn becomes an entry. Storing it turns reconciliation from "work out what this was" into "confirm what it probably was".

## What Changes

- **BREAKING** Usage entries record the branch they were captured on. `usage_entries` gains a `branch` column and `usage.Entry` a `Branch` field; the capture path stops discarding what it already reads.
- `usage.UnattributedGroup` regains `Branch` and gains `Suggested`, an issue number derived from the branch, populated from stored data rather than left nil.
- `tokenpile unattributed` shows the branch and the suggested issue, and `assign` accepts the suggestion without repeating the number.
- New TUI view listing unattributed groups: repository, branch, time window, entries, tokens, cost, and the suggested issue. Assigning and unassigning happen in the view, with the suggestion pre-filled and editable.
- The view is reachable from the issue list, and surfaces itself when unattributed usage exists, so it is discovered rather than looked up.

## Capabilities

### Modified Capabilities

- `attribution`: entries record their branch; reconciliation offers a branch-derived suggestion, which `deterministic-capture` explicitly deferred for want of stored data.
- `tui`: a new view for reviewing and assigning unattributed usage.

## Impact

- `internal/usage/types.go`: `Entry.Branch`; `UnattributedGroup` gains `Branch` and `Suggested`.
- `internal/store/sqlite.go`: `branch` column, insert and scans, grouping by branch alongside session, suggestion derived on read.
- `cmd/tokenpile/cmd_hook.go`: carry `Turn.Branch` through to the entry instead of using it only for attribution.
- `cmd/tokenpile/cmd_bind.go`: show branch and suggestion; `assign` can take the suggestion.
- `internal/tui/`: new view, key bindings, and its place in the navigation.
- No pricing, export or capture-format change.
- Depends on `deterministic-capture`, which introduced unattributed usage and the CLI this builds on.
