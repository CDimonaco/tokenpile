## Context

All local state lives in four places: the data dir (`tokenpile.db` plus possible `-wal`/`-shm` sidecars), the config dir (`identity.key`, `identity.pub`, `credentials`, `pricing.yaml`), the OS keychain (`tokenpile`/`github-token`), and agent directories (`~/.claude/skills/tokenpile.md`, marked blocks in codex/opencode `AGENTS.md`). `config.Paths` already resolves the first two. `provider.Logout` already removes keychain token and credentials file. `skill.Install` exists but has no inverse. Export currently gathers sessions and budgets only when the filter names a single repo+issue (`cmd_export.go`).

## Goals / Non-Goals

**Goals:**
- One command that leaves the machine as if tokenpile had never run, minus the backup file.
- The backup is a standard signed export document (schema 3.0), complete: all entries, all sessions, all budgets.
- Safe by default: explicit confirmation, dump before destruction, key deleted only after the dump is signed.

**Non-Goals:**
- Import/restore. The dump format is the standard export document precisely so a future `import` command can consume it; nothing else is done now.
- Uninstalling the binary (that is `make uninstall`).
- Selective resets (`--auth-only` etc.); full reset only.

## Decisions

**1. Export scope rule: sessions and budgets follow the entries repo/issue filter.**
Unfiltered → all sessions and budgets; `--repo` → that repo's; `--repo --issue` → unchanged. Time/agent/model filters do not apply to sessions and budgets (sessions are not attributable to a model, and a truncated session list would misrepresent wall-clock time). Alternative considered: a reset-private dump path — rejected, two serialization paths to maintain and the export limitation would keep biting other users.

**2. New store methods instead of filter overloading.**
`ListAllSessions(ctx)` and `ListBudgets(ctx)` are plain queries with no WHERE clause (plus repo-scoped variants via the existing pattern if needed by the CLI). `ListSessions(ctx, repo, issueNum)` stays as is; widening its signature would touch every call site for no benefit.

**3. Reset execution order.**
1. Resolve paths and enumerate what exists (DB, keys, credentials, pricing override, keychain entry, installed skills).
2. Print the list; prompt for `yes` unless `--yes`.
3. Build and write the signed dump (unless `--no-backup`). Abort on any dump error: no destruction without a successful backup when one was requested.
4. Destroy: keychain token + credentials (via `Logout`), skill files, pricing override, identity keypair, DB and sidecar files. Deleting the open SQLite file is safe on unix (unlink); the store is closed by main's deferred Close.
5. Print what was removed and the backup path.
Deletion failures after a successful dump do not abort the remaining deletions: collect errors, report all, exit non-zero if any.

**4. `skill.Uninstall(agentName)` mirrors `Install`.**
Dedicated file: remove it. Shared file: remove the `<!-- tokenpile:start -->`/`<!-- tokenpile:end -->` block and surrounding blank padding; if the file becomes empty, remove it; if no marker, report not installed (no error). Returns (path, removed bool, error).

**5. Confirmation reads from stdin, not a TTY library.**
`fmt.Fscanln` on `cli.App.Reader`-equivalent (os.Stdin) keeps it testable by injecting a reader; exact string `yes` required, anything else aborts with exit code 1 and no side effects.

**6. Backup file naming.**
`tokenpile-backup-20060102-150405.json` in the working directory, 0600 like other exports. `--output -` is not supported; reset is inherently file-oriented.

## Risks / Trade-offs

- [Unfiltered exports grow: every session and budget in the DB] → acceptable for a local tool; the arrays were already unbounded per issue.
- [External consumers surprised by populated `sessions`/`budgets` in unfiltered exports] → schema shape is unchanged (arrays already exist, were empty); called out as BREAKING in the proposal anyway.
- [Reset deletes the identity key: origin verification of old backups against a "current" key becomes impossible] → the dump embeds its public key; consistency verification works forever, and the user can archive `identity.pub` separately before reset if provenance matters.
- [Partial deletion failure leaves half-reset state] → errors are collected and reported per item; the command is idempotent, rerunning finishes the job.

## Migration Plan

Implement export scope change first (with tests), then skill.Uninstall, then the reset command on top. Single change, sequential tasks. Rollback: revert; no data formats change shape.

## Open Questions

None.
