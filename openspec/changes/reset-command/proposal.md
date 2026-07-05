## Why

There is no way to cleanly wipe tokenpile from a machine: tokens live in the keychain, keys and credentials in the config dir, usage data in the data dir, and skill files inside third-party agent directories. Removing them by hand is error-prone and loses data silently. A `reset` command must archive everything in a signed dump first, then destroy all local state in one step. Today's export cannot serve as that dump: sessions and budgets are only included when filtering a single repo+issue, so an unfiltered backup would lose notes, tags, session times and budgets.

## What Changes

- New `tokenpile reset` command: interactive confirmation listing exactly what will be deleted (type `yes`; `--yes` skips the prompt for scripts), then a signed export dump, then destruction of all local state.
- Dump: written to `./tokenpile-backup-<timestamp>.json` by default, `--output` overrides the path, `--no-backup` skips it. The dump MUST be produced before the identity key is deleted, since the key signs it.
- Destruction scope: SQLite DB plus WAL/SHM sidecar files, Ed25519 identity keypair, encrypted credentials file, keychain token, `pricing.yaml` override, and installed agent skills (dedicated file for claude-code; marked-block removal from shared `AGENTS.md` for codex and opencode).
- New `skill.Uninstall` in `internal/skill`: removes the dedicated file or strips the marked block from shared files.
- **BREAKING** Export scope rule change: sessions and budgets follow the same repo/issue filter as entries. Unfiltered export includes ALL sessions and budgets; `--repo` includes the whole repo; `--repo --issue` behaves as today. Requires new store methods `ListAllSessions` and `ListBudgets`.
- No import command yet, but the dump format (the standard signed export document) stays importable for a future restore command.

## Capabilities

### New Capabilities

- `reset`: the `tokenpile reset` command — confirmation flow, pre-destruction signed dump, full deletion scope, ordering guarantees.

### Modified Capabilities

- `export`: the "Signed JSON export format" requirement changes — sessions and budgets are scoped by the entries filter instead of appearing only for single repo+issue exports.
- `agent-skill`: new requirement — skills can be uninstalled (dedicated file removal, marked-block removal in shared files).

## Impact

- `cmd/tokenpile/`: new `cmd_reset.go` wired in `main.go`; reset needs store, paths, auth provider, identity key and version.
- `internal/store/`: `ListAllSessions(ctx)` and `ListBudgets(ctx)` added to the `Store` interface and SQLite adapter; mocks regenerated (`make generate`).
- `cmd/tokenpile/cmd_export.go`: session/budget gathering follows the entries filter.
- `internal/skill/skill.go`: `Uninstall(agentName)`.
- `internal/provider/`: reuse `Logout` for token/credentials removal.
- Tests: store, skill, export integration, reset integration (fresh temp dirs), smoke.
- Docs: README (`reset` section, export scope), CLAUDE.md project map (new command file).
- External consumers of unfiltered exports will start seeing `sessions` and `budgets` arrays.
