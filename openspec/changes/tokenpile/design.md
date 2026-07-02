## Context

tokenpile is a greenfield Go CLI/TUI tool. There is no existing codebase to migrate. The tool runs locally on developer machines (macOS and Linux), stores all data in a local SQLite database, and is invoked both by humans (TUI) and by LLM agents (CLI subcommands). Cross-platform binary distribution and zero-infrastructure operation are hard constraints.

## Goals / Non-Goals

**Goals:**
- Track LLM token usage, cost, and wall-clock time per GitHub issue across any agent
- Provide a terminal UI for humans and a clean CLI surface for agents
- Support GitHub as the first issue provider, with other providers possible via interfaces
- Ship as a single static binary for macOS and Linux via goreleaser and Homebrew
- All project conventions (style, testing, tooling) encoded in CLAUDE.md so agents working on this repo stay consistent

**Non-Goals:**
- Real-time sync across multiple machines in v1
- Web dashboard or hosted backend
- Team/multi-user aggregation in v1
- Container image (OS keychain and TUI are incompatible with typical container use)
- Cached token tracking (not modeled; raw in/out only)

## Decisions

### Language and CLI framework: Go + urfave/cli

Go produces a single static binary, cross-compiles easily, and has excellent TUI library support. `urfave/cli` is chosen over cobra for a lighter surface — tokenpile does not need the plugin ecosystem cobra provides.

Alternatives considered: Rust (steeper learning curve, slower iteration), Node/Deno (no clean single-binary story).

### TUI: Bubble Tea + lipgloss + ntcharts

Bubble Tea is the standard Go TUI framework. `ntcharts` is chosen for chart views (usage over time) because it integrates natively with Bubble Tea's model/update/view pattern. `lipgloss` handles layout and styling.

### Storage: modernc.org/sqlite (no CGO)

`modernc.org/sqlite` is a pure-Go SQLite implementation. It eliminates CGO, making cross-compilation to Linux from macOS trivial without a C toolchain. Performance difference over `mattn/go-sqlite3` is negligible for this workload.

The `Store` interface decouples all domain logic from SQLite. Future adapters (e.g., remote sync, Postgres) can be plugged in without touching domain or CLI code.

### Auth and issue providers: interface-first

Both `AuthProvider` and `IssueProvider` are interfaces. GitHub is the first implementation. This allows future providers (GitLab, Linear, Jira) to be added without changing call sites.

OAuth uses a local callback server (not device flow). A local HTTP server on an ephemeral port receives the OAuth redirect, exchanges the code, and stores the token. This strategy is identical across providers and works reliably on both macOS and Linux with a desktop environment.

### Token storage: OS keychain via zalando/go-keyring

OAuth tokens grant external access and must be protected. `zalando/go-keyring` uses macOS Keychain on macOS and the Secret Service (libsecret/D-Bus) on Linux. On headless Linux without a Secret Service, it falls back to an AES-encrypted file at `~/.config/tokenpile/credentials` (0600).

### Signing keypair: Ed25519 files at 0600

The signing keypair is NOT stored in the keychain. It is a local identity analogous to an SSH key. The private key lives at `~/.config/tokenpile/identity.key` (0600) and the public key at `~/.config/tokenpile/identity.pub` (0644). Both are generated on first run if absent. Storing in keychain would trigger OS prompts on every export operation, which is poor UX.

### Cost: computed at report time, never stored

`UsageEntry` stores raw token counts and model name only. Cost is derived at report/export time by looking up the model in the merged pricing config. This means pricing corrections (model price changes, new models) are reflected in historical reports without re-logging.

### Sessions: implicit, idle-timeout based

Sessions are not managed explicitly by the user or agent. A session starts automatically on the first `tokenpile log` call for an issue when no session is active. A background ticker closes sessions idle for more than 30 minutes. This gives wall-clock time without requiring agents to call `session start/end`.

### Package structure

```
cmd/tokenpile/        - main.go; wires CLI, injects dependencies
internal/
  domain/             - shared types: UsageEntry, Issue, Session, Report
  store/              - Store interface + SQLite adapter
  provider/           - AuthProvider, IssueProvider interfaces + GitHub impl
  pricing/            - config loading, two-layer merge, cost computation
  export/             - JSON marshaling, Ed25519 signing, schema validation
  skill/              - embedded skill templates, install logic
  tui/                - Bubble Tea app, views (list, detail, charts)
  config/             - XDG-compliant path resolution
```

### Dependency injection: constructor functions only

All dependencies are injected via constructor functions. No DI library is used. `main.go` owns the composition root: it reads config, opens the DB, constructs adapters, and passes them into CLI commands and TUI. This keeps the dependency graph explicit and testable.

### Error handling

Errors at package boundaries use `fmt.Errorf("operation: %w", err)` for context wrapping. Sentinel errors (e.g., `ErrNoRepo`, `ErrSessionNotFound`) are defined for errors callers need to match. Custom error types are used only when structured data is needed (e.g., `RepoInferError` to include the path that was searched).

### Logging: slog with optional JSON mode

All internal logging uses `log/slog`. The default handler is text. Passing `--log-format json` (or setting `TOKENPILE_LOG_FORMAT=json`) switches to `slog.NewJSONHandler`. Log level is controlled via `--log-level` or `TOKENPILE_LOG_LEVEL`. No logging frameworks beyond stdlib.

### Testing conventions

- Unit tests: `stdlib testing` + `testify/assert` + `testify/require`
- Mocks: generated by `mockery` from interfaces
- Integration tests: each test gets its own SQLite DB (temp file, cleaned up via `t.Cleanup`); no shared state between test cases
- Linting: `golangci-lint` with the maratori golden config
- `context.Context` is the first parameter of any function that performs I/O or is cancelable

### Dev tooling: make + asdf-vm

`make` is the task runner. The `Makefile` provides a single entry point for all common developer operations:
- `make build` — `go build ./cmd/tokenpile`
- `make test` — `go test -race ./...`
- `make lint` — `golangci-lint run --timeout 5m`
- `make fmt` — `gofmt -w .`
- `make generate` — regenerate mockery mocks
- `make install` — install binary to `$GOPATH/bin`
- `make clean` — remove build artifacts
- `make release-check` — `goreleaser check` (validate release config without publishing)

`asdf-vm` manages all dev dependency versions via `.tool-versions` at the repo root. Pinned tools: `golang`, `golangci-lint`, `goreleaser`, `mockery`. CI reads the same `.tool-versions` file via the `asdf` GitHub Action so local and CI environments are identical.

### Commits and authorship

Conventional commits: `feat:`, `fix:`, `chore:`, `docs:`, `test:`, `refactor:`, `ci:`. Never include AI agent co-author trailers. Commits carry only the human author's identity.

### Release: goreleaser + Homebrew

goreleaser builds and publishes binaries for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`. A Homebrew formula is pushed to `cdimonaco/homebrew-tokenpile` on each release tag. Checksums are published alongside binaries.

## Risks / Trade-offs

OS keychain on headless Linux → Mitigation: encrypted file fallback; document clearly in README.

Local callback OAuth requires a browser → Mitigation: document that headless-only environments must set up SSH tunneling or use a machine with a browser for initial auth; token refresh is handled silently after initial login.

Ed25519 keypair is machine-scoped → Mitigation: `tokenpile identity export/import` can be added in a future change; document the limitation at export time.

Pricing config can drift from reality → Mitigation: `tokenpile pricing list` shows current config; warn in reports when a model is missing from pricing.

ntcharts dependency maturity → Mitigation: charts are isolated in the `tui/` package; if the library proves inadequate it can be swapped without touching domain logic.

## Open Questions

None. All design decisions were settled during the exploration phase.
