## 1. Store: full listings

- [x] 1.1 Add `ListAllSessions(ctx) ([]usage.Session, error)` and `ListBudgets(ctx) ([]usage.IssueBudgetRow, error)` (repo, issue_num, budget) to the `Store` interface and the SQLite adapter; sessions query has no WHERE, budgets reads `issue_budgets`
- [x] 1.2 Run `make generate` to regenerate mocks
- [x] 1.3 Store tests: sessions across repos returned in full; budgets listing round-trips `SetBudget`

## 2. Export scope change

- [ ] 2.1 Rework session/budget gathering in `runExport` (`cmd/tokenpile/cmd_export.go`): no filter → all sessions and budgets; `--repo` → whole repo; `--repo --issue` → unchanged; time/agent/model filters do not apply to sessions and budgets
- [ ] 2.2 Update `TestIntegration_Export_NoRepoIssueFilter_OmitsSessionsAndBudgets` (now expects inclusion) and add repo-scoped and unfiltered integration tests matching the spec scenarios

## 3. Skill uninstall

- [ ] 3.1 Implement `skill.Uninstall(agentName) (string, bool, error)`: dedicated file removed; shared file loses only the marked block (file removed if nothing else remains); not-installed is success with removed=false
- [ ] 3.2 Skill tests: dedicated removal, shared removal preserving foreign content, empty-after-removal deletes file, not-installed no-op

## 4. Reset command

- [ ] 4.1 New `cmd/tokenpile/cmd_reset.go`: enumerate existing state (DB + `-wal`/`-shm`, identity keypair, credentials, pricing override, keychain entry, installed skills), print the deletion list, prompt for `yes` on stdin unless `--yes`
- [ ] 4.2 Backup step: build the full export (all entries, sessions, budgets) signed with the identity key, write to `tokenpile-backup-<timestamp>.json` (0600), `--output` override, `--no-backup` skip; abort everything if the backup write fails
- [ ] 4.3 Destruction step: `Logout` (keychain + credentials), `skill.Uninstall` for every agent, remove pricing override, identity keypair, DB and sidecars; treat missing items as already removed; collect per-item errors, report all, exit non-zero if any
- [ ] 4.4 Wire `resetCommand` in `main.go` with store, paths, auth provider, private key and version
- [ ] 4.5 Integration tests with temp dirs: confirmation abort leaves everything intact; `--yes` resets and backup verifies with `export verify`; `--no-backup` skips the file; second run succeeds; backup failure aborts destruction

## 5. Docs and checks

- [ ] 5.1 README: `tokenpile reset` section and updated export scope wording; CLAUDE.md project map entry for `cmd_reset.go`
- [ ] 5.2 Run `make check` and commit
