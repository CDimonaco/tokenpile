## 1. Store the branch

- [ ] 1.1 Add `Branch` to `usage.Entry` and a `branch` column to `usage_entries`; update the insert and every scan
- [ ] 1.2 `cmd_hook.go`: carry `Turn.Branch` into the entry instead of using it only for attribution
- [ ] 1.3 Store tests: branch round-trips, an absent branch reads back empty

## 2. Suggestions

- [ ] 2.1 Restore `Branch` on `usage.UnattributedGroup` and add `Suggested *int`, derived on read via `attribution.InferFromBranch`
- [ ] 2.2 Group by session and branch in `ListUnattributed`; carry the time window already computed
- [ ] 2.3 `tokenpile unattributed` shows branch and suggestion; `assign` accepts the suggestion without repeating the number
- [ ] 2.4 Tests: a group from `feat/42-x` suggests 42, a group from `main` suggests nothing, pre-existing entries with no branch suggest nothing

## 3. TUI view

- [ ] 3.1 New view listing groups: repository, branch, time window, entries, tokens, cost, suggestion
- [ ] 3.2 Assign action with the suggestion pre-filled and editable; unassign action; both delegating to the existing store methods
- [ ] 3.3 Empty state that says so rather than rendering an empty table
- [ ] 3.4 Issue list indicates unattributed usage exists and offers the key to open the view
- [ ] 3.5 TUI tests: list rendering, suggestion pre-fill, assigning removes the group, empty state, the indicator appears only when warranted

## 4. Docs and checks

- [ ] 4.1 README: the unattributed view, its key bindings, and what a suggestion means
- [ ] 4.2 CLAUDE.md: note that the branch is stored at capture time and never re-derived
- [ ] 4.3 Run `make check` and commit
