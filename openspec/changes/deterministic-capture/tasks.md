## 1. Nullable issue

- [x] 1.1 `Entry.IssueNum` and `Session.IssueNum` become `*int` in `internal/usage/types.go`
- [x] 1.2 `usage_entries.issue_num` and `sessions.issue_num` become nullable; update inserts and scans
- [x] 1.3 Update every query filtering or grouping by `issue_num` to handle null explicitly rather than silently excluding rows
- [x] 1.4 Store methods for unattributed listings and for assigning/unassigning a group of entries
- [x] 1.5 `make generate` for store mocks
- [x] 1.6 Store tests: unattributed entries persist and read back, aggregations do not drop them, assignment and reversal round-trip

## 2. Capture readers

- [ ] 2.1 New `internal/capture/`: the internal per-turn shape (per-tier usage, model, agent, session id, cwd, timestamp)
- [ ] 2.2 Claude Code reader: parse the JSONL transcript, take `message.usage` and `message.model` from assistant messages, map `input_tokens`/`cache_creation_input_tokens`/`cache_read_input_tokens`/`output_tokens` onto the tiers, reasoning zero
- [ ] 2.3 opencode reader: map `tokens.{input,output,reasoning,cache.{write,read}}` and `modelID` onto the tiers
- [ ] 2.4 Reader tests against captured fixtures of both formats, including a turn with no usage block and a malformed line

## 3. Spool and reconciliation

- [ ] 3.1 Append-only spool writer: one JSON record per turn, created without needing the database
- [ ] 3.2 Reconciler folding the spool into the store, idempotent on records already applied
- [ ] 3.3 Run reconciliation on tokenpile invocations; surface the backlog in `report` output
- [ ] 3.4 Unparseable payloads are spooled raw and reported, never dropped
- [ ] 3.5 Tests: database unavailable still spools, draining twice creates no duplicates, malformed payload preserved

## 4. Attribution

- [ ] 4.1 New `internal/attribution/`: binding store keyed by session id and working directory
- [ ] 4.2 Offline branch-name inference with documented patterns; no network calls
- [ ] 4.3 Resolution order at capture time: binding, then branch, then none
- [ ] 4.4 `cmd_bind.go`: `tokenpile bind --issue N [--note --tag]`; rebinding replaces
- [ ] 4.5 Tests: binding wins over branch, branch inference patterns, `main` yields nothing and makes no network call, rebinding does not retroactively move earlier entries

## 5. Hook and plugin installation

- [ ] 5.1 `tokenpile hook claude-code`: read hook JSON from stdin, resolve the transcript, spool the turn
- [ ] 5.2 Install a `Stop` hook into the Claude Code settings file by merging, preserving all foreign hooks and settings, with a marker identifying tokenpile's entry
- [ ] 5.3 Embed and install the opencode plugin subscribing to `session.idle`
- [ ] 5.4 `skill uninstall` and `reset` remove the hook entry and the plugin, leaving foreign settings intact
- [ ] 5.5 Tests: install into a settings file containing unrelated hooks, uninstall leaves them, reinstall is idempotent

## 6. Skill template

- [ ] 6.1 Rewrite the remaining templates: remove all token estimation instructions, add `tokenpile bind`, keep the usage-question responsibilities; bump the skill version marker
- [ ] 6.2 Test asserting no template mentions token counts

## 7. Reporting unattributed usage

- [ ] 7.1 Decide and implement how unattributed usage appears in `report` and in the TUI, so it is never invisible spending
- [ ] 7.2 Budgets are per issue: confirm and document that unattributed usage counts toward no budget, and make that visible
- [ ] 7.3 TUI view listing unattributed groups by repo, branch and time window with suggestions and bulk assignment

## 8. Docs and checks

- [ ] 8.1 README: capture, bind, reconciliation, what installing a skill now installs
- [ ] 8.2 CLAUDE.md: project map for `internal/capture/` and `internal/attribution/`; design decisions gain nullable attribution and the spool
- [ ] 8.3 Run `make check` and commit
