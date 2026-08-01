## 1. Store the branch

- [x] 1.1 Add `Branch` to `usage.Entry` and a `branch` column to `usage_entries`; update the insert and every scan
- [x] 1.2 `cmd_hook.go`: carry `Turn.Branch` into the entry instead of using it only for attribution
- [x] 1.3 Store tests: branch round-trips, an absent branch reads back empty

## 2. Suggestions

- [x] 2.1 Restore `Branch` on `usage.UnattributedGroup` and add `Suggested *int`, derived on read via `attribution.InferFromBranch`
- [x] 2.2 Group by session and branch in `ListUnattributed`; carry the time window already computed
- [x] 2.3 `tokenpile unattributed` shows branch and suggestion; `assign` accepts the suggestion without repeating the number
- [x] 2.4 Tests: a group from `feat/42-x` suggests 42, a group from `main` suggests nothing, pre-existing entries with no branch suggest nothing

## 3. TUI view

- [x] 3.1 New view listing groups: repository, branch, time window, entries, tokens, cost, suggestion
- [x] 3.2 Assign action with the suggestion pre-filled and editable; unassign action; both delegating to the existing store methods
- [x] 3.3 Empty state that says so rather than rendering an empty table
- [x] 3.4 Issue list indicates unattributed usage exists and offers the key to open the view
- [x] 3.5 TUI tests: list rendering, suggestion pre-fill, assigning removes the group, empty state, the indicator appears only when warranted

## 4. Docs and checks

- [x] 4.1 README: the unattributed view, its key bindings, and what a suggestion means
- [x] 4.2 CLAUDE.md: note that the branch is stored at capture time and never re-derived
- [x] 4.3 Run `make check` and commit
