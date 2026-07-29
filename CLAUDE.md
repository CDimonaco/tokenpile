# tokenpile — project conventions

## Working with AI agents

These rules apply when an AI agent (Claude Code or similar) is implementing changes in this repo. They exist so the developer can follow along, intervene, and review without being in yolo mode.

### Plan before code

For any change that touches more than one package or introduces a new concept: write a 3-6 line plan in the conversation before editing any files, and wait for explicit approval. One-liner fixes and isolated test additions do not need a plan.

### Commit rhythm

Commit after each completed task — not at the end of a session. Run `make check` before every commit. This gives the developer a reviewable, revertable checkpoint at every step.

### Announce before large changes

Before modifying more than three files or touching a package for the first time in a session, state in one sentence what is about to change and why. Then start the edits.

### Never push without being asked

Do not run `git push` unless the developer explicitly requests it in that message. Implement and commit freely; the developer decides when to publish.

## OpenSpec

Planning artifacts live in `openspec/`. Two different things get validated, and passing one says nothing about the other:

- `openspec validate <change>` — checks a change's **delta** specs in `openspec/changes/<name>/specs/`
- `openspec validate --specs` — checks the **main** specs in `openspec/specs/`

Run both. `make check` runs `--specs` via the `spec-check` target, so it gates every commit.

### Main specs vs delta specs have different shapes

They are not interchangeable, and mixing them silently destroys content.

A **delta** spec (inside a change) is organised by operation:

```markdown
## ADDED Requirements
### Requirement: ...
## MODIFIED Requirements
### Requirement: ...
```

A **main** spec (in `openspec/specs/<capability>/spec.md`) is a flat document:

```markdown
## Purpose

[what this capability is for]

## Requirements

### Requirement: ...
#### Scenario: ...
```

Rules when syncing a delta into a main spec:

- Never copy `## ADDED Requirements` / `## MODIFIED Requirements` / `## REMOVED Requirements` headers into a main spec. A stray `##` header **truncates the requirements section**: everything below it stops being parsed, so the capability silently reads as having zero requirements. This happened to four specs and went unnoticed for weeks.
- Every main spec must open with `## Purpose` and `## Requirements`. A sync appends to an existing file and will not create this skeleton for you — check it is there.
- A `MODIFIED` requirement replaces the one of the same name. Before writing one, read what the main spec already has: scenarios you fail to carry over are scenarios you delete.
- After any sync or archive, run `openspec validate --specs` and read the output. `Spec must have at least one requirement` on a file that visibly contains requirements means the section is truncated, not empty.

## Commits

- Conventional commits: `feat:`, `fix:`, `chore:`, `docs:`, `test:`, `refactor:`, `ci:`
- Atomic commits: one logical change per commit; never batch unrelated work into a single commit.
- Never add AI agent co-author trailers. Commits carry only the human author's identity.
- No "Co-Authored-By" lines of any kind.

## No emojis

Never use emojis anywhere in code, comments, commit messages, or output strings.

## Project map

Module: `github.com/cdimonaco/tokenpile`

A machine-generated call graph is committed at `.gograph/`. Read it before using `ls` or `find`:

- `.gograph/GRAPH_REPORT.md` — entry point: package count, symbol count, entry points, index of other reports
- `.gograph/graph-symbols.md` — top files by density, top symbols by outgoing calls, package table
- `.gograph/graph-deps.md` — tech stack (go.mod) and per-package import lists
- `.gograph/graph-errors.md` — custom errors and panics with origin file:line
- `.gograph/graph-sql.md` — raw SQL queries mapped to functions

The graph is regenerated on every `make check` (run before each commit), so it is always current.
Do NOT read `.gograph/graph.json` — it is a large binary database that will exhaust context.

Key files — read these directly without exploring the directory first:

```
cmd/tokenpile/
  main.go                 composition root; wires config, store, provider, commands
  cmd_log.go              `log` command
  cmd_report.go           `report` command
  cmd_export.go           `export` command
  cmd_budget.go           `budget` command
  cmd_auth.go             `auth` command (OAuth login/logout)
  cmd_pricing.go          `pricing` command
  cmd_skill.go            `skill` command
  cmd_hook.go             `hook` command (agent-invoked) + spool reconciliation
  cmd_bind.go             `bind` and `unattributed` commands
  cmd_reset.go            `reset` command (signed backup, then full state deletion)
  integration_test.go     CLI integration tests; helpers: newTestStore, runLogCmd, runReportCmd, runBudgetCmd, runExportCmd

internal/
  usage/
    usage.go              Entry, Session, Filter, Report, TrackedIssue, IssueCache, Point
  store/
    store.go              Store interface (all DB operations)
    sqlite.go             SQLite adapter (modernc.org/sqlite, no CGO)
  provider/
    provider.go           AuthProvider, IssueProvider interfaces; Issue type
    github_auth.go        OAuth flow implementation
    ghcli_auth.go         gh CLI as an alternative token source
    tokensource.go        which token source this machine uses (persisted choice)
    validate.go           token validation (login, scopes) before committing to a credential
    github_issues.go      GitHub Issues API client
    repoinfer.go          infer repo from git remote
  pricing/
    pricing.go            two-layer config loader + cost computation
    pricing.defaults.yaml embedded default pricing
  export/
    export.go             JSON export, Ed25519 signing, schema validation
  capture/
    capture.go            Turn: one captured unit of work, normalized across agents
    claudecode.go         Stop-hook payload + JSONL transcript reader
    opencode.go           session.idle plugin payload reader
    spool.go              append-only journal so a failing hook loses nothing
  attribution/
    attribution.go        issue bindings, offline branch inference, resolution order
  skill/
    skill.go              List, Install, Uninstall, IsInstalled, IsUpToDate
    hooks.go              capture hook install/removal (settings.json merge, opencode plugin)
    templates/            embedded agent skill files
  tui/
    tui.go                Bubble Tea Model, Update, View; all TUI views
  config/
    paths.go              XDG path resolution
    identity.go           Ed25519 key generation and loading
```

## Files to skip unless investigating a specific error

Do not read these proactively — they are managed by tooling and are never edited by hand:

- `go.mod`, `go.sum` — use `go get` / `go mod tidy`; the module path is `github.com/cdimonaco/tokenpile`
- `internal/*/mocks/` — generated by `make generate` (mockery); never edit, read only to inspect interface shape
- `./tokenpile` — built binary, never read

## Package structure

```
cmd/tokenpile/        main.go — composition root only
internal/
  usage/              shared types: Entry, Session, Filter, Report, TrackedIssue, Point
  store/              Store interface + SQLite adapter
  provider/           AuthProvider, IssueProvider interfaces + Issue type + GitHub implementations
  pricing/            two-layer pricing config, cost computation
  export/             JSON marshaling, Ed25519 signing, schema validation
  skill/              embedded skill templates, install logic
  tui/                Bubble Tea app, views (list, detail, charts)
  config/             XDG-compliant path resolution, Ed25519 identity management
```

## Package naming

Name packages after what they contain, not architectural layers. No `domain`, `model`, or `dto` packages — those are Java patterns and add indirection without meaning.

- Shared types that describe the application's core data live in `internal/usage` — named after the concept, not the layer.
- Types that originate from a specific provider live in that provider's package (e.g. `provider.Issue`).
- When a type name would repeat the package name, drop the prefix: `usage.Entry` not `usage.UsageEntry`, `usage.Point` not `usage.UsagePoint`.

## Dependency injection

- Inject all dependencies via constructor functions. No DI library.
- `main.go` is the composition root: reads config, opens DB, constructs adapters, passes them in.
- Keep the dependency graph explicit. If a function needs something, it receives it as a parameter.

## Error handling

- Wrap errors at package boundaries: `fmt.Errorf("operation: %w", err)`
- Define sentinel errors for values callers need to match: `var ErrNoRepo = errors.New("...")`
- Custom error types only when structured data is needed alongside the error.
- Do not add error handling for scenarios that cannot happen.

## Logging

- Use `log/slog` only. No third-party logging libraries.
- Default handler: text. JSON mode: enabled via `--log-format json` or `TOKENPILE_LOG_FORMAT=json`.
- Log level: `--log-level` flag or `TOKENPILE_LOG_LEVEL` env var (debug/info/warn/error, default info).
- Agents and machines should use `--log-format json` to parse output.

## Context

- `context.Context` is the first parameter of any function that performs I/O or is cancelable.

## Interfaces

- Define sync interfaces first. Wrap in async/concurrent only when needed.
- Keep interfaces small and focused. Define them where they are consumed, not where implemented.

## Testing

- Unit tests: `stdlib testing` + `testify/assert` + `testify/require`
- Mocks: generated by `mockery`. Run `make generate` to regenerate.
- Integration tests: each test case gets its own fresh SQLite DB (temp file via `t.TempDir()`), cleaned up with `t.Cleanup`. No shared state between test cases.
- Each test must be independent and self-contained.
- Test files live alongside the code they test (`*_test.go` in the same package for white-box, `package foo_test` for black-box).

## Linting

- `golangci-lint` with `.golangci.yml` (maratori golden config).
- Run: `make lint`
- CI enforces lint on every push.

## Formatting

- `gofmt -w .` — run via `make fmt`
- CI fails on unformatted files.

## Dev tooling

- `asdf-vm` manages all tool versions via `.tool-versions` at the repo root.
- Set up: `asdf install` installs all pinned versions.
- `make` is the task runner. Always use make targets; do not invoke tools directly in CI.

## Make targets

| Target          | What it does                        |
|-----------------|-------------------------------------|
| `make build`    | Build binary to `./tokenpile`       |
| `make test`     | Run all tests with race detector    |
| `make lint`     | Run golangci-lint                   |
| `make fmt`      | Format all Go files                 |
| `make generate` | Regenerate mocks via mockery        |
| `make install`  | Install binary to `$GOPATH/bin`     |
| `make uninstall` | Remove the binary installed by `make install` |
| `make clean`    | Remove build artifacts              |
| `make release-check` | Validate goreleaser config     |
| `make check`    | fmt + lint + test + spec-check + map (run before commit) |
| `make spec-check` | Validate main specs (`openspec validate --specs`) |
| `make status`   | Branch, uncommitted files, recent commits, OpenSpec changes, CI status |
| `make map`      | Regenerate `.gograph/` call graph (gograph) |
| `make pack`     | Compress full source to `.gograph/pack.md` (repomix, on demand) |
| `make tools`    | Install dev tools: gograph |

## Key design decisions (summary)

- Storage: `modernc.org/sqlite` (no CGO, easy cross-compilation)
- CLI: `urfave/cli/v2`
- TUI: Bubble Tea + lipgloss + ntcharts
- OAuth: local callback server, not device flow. Ephemeral loopback port (GitHub ignores the port on loopback redirects) plus PKCE (S256) so a captured authorization code cannot be exchanged. The OAuth client secret is embedded in release binaries: GitHub requires it at token exchange even with PKCE. Accepted limitation — with PKCE and an ephemeral port the extractable secret only allows app impersonation, not token theft.
- OAuth tokens: OS keychain via `zalando/go-keyring`; headless Linux falls back to AES-256-GCM encrypted file
- Token source: the OAuth App, or the credential held by the `gh` CLI, for organizations that never approve the OAuth App. `gh` is a token source only — issue calls still go through go-github. The choice is made once at `auth login` and persisted as a `gh-cli:` sentinel in the existing credential slot (no config file); never a per-call decision and never a silent fallback, since a signed-export tool must not be ambiguous about which credential answered. The borrowed token is never cached: `gh auth token` runs on every call so expiry stays with `gh`.
- Signing keypair: Ed25519 files at `~/.config/tokenpile/identity.{key,pub}` (0600/0644), generated on first run
- Export signature (schema 4.0): covers the canonical JSON of the whole document with the `signature` field emptied. Only 4.0 verifies — the digest is taken over the *parsed* document, so it depends on the current field set and no earlier schema can verify without a per-version parsing type. `export verify --pubkey` checks the embedded key against an expected key to prove origin; without it, verification proves internal consistency only.
- Cost: computed at report time from pricing config, never stored
- Token tiers: usage is recorded as `Usage{InputFresh, CacheWrite, CacheRead, Output, Reasoning}`, not two buckets. Providers bill these very differently (cached reads at ~10% of input, cache writes above it), so two buckets cannot represent the bill. `Reasoning` is a SUBSET of `Output` and must never be added on top of it. Pricing declares four explicit rates per model; a tier with tokens but no rate warns and is excluded from cost rather than being charged at the input rate or an assumed ratio.
- Entry provenance: every entry carries `source` of `measured` or `estimated`, derived from the command that writes it, never from a flag. An agent cannot observe its own cache tiers, so the distinction is structural.
- Capture: token counts come from the agent's own transcript, never from a model's estimate. Claude Code fires a `Stop` hook (in `settings.json`, NOT skill frontmatter — frontmatter hooks are scoped to the skill's lifecycle and only run while the model chose to load it, which is the non-determinism capture exists to remove); opencode fires `session.idle` to an installed plugin. No agent payload carries token counts: the hook says when and where, the transcript says how much.
- Attribution: `issue_num` is nullable on entries and sessions. That constraint was the root cause of the estimates — nothing could be recorded without an issue, only the model knew the issue, so the model had to log and therefore had to invent the numbers. Resolution order is binding, then offline branch-name inference, then none. A missing attribution is an unattributed measurement, never a lost one, and shows in `report` because it counts toward no budget.
- Spool: capture appends to an append-only journal before any database write, and reconciliation folds it in on later invocations. A hook exiting non-zero merely prints to stderr and continues, so writing straight to SQLite would drop turns silently. Storage is idempotent on the entry id (`INSERT OR IGNORE`), which is what makes clearing the spool safe without a two-phase commit.
- Sessions: implicit, 30-minute idle auto-close
- Repo: inferred from `git remote get-url origin` when not passed explicitly; fails with clear error if not inferable
