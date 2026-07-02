## 1. Project Bootstrap

- [x] 1.1 Initialize Go module `github.com/cdimonaco/tokenpile` with `go mod init`
- [x] 1.2 Create `.tool-versions` pinning: `golang`, `golangci-lint`, `goreleaser`, `mockery` at their current stable versions; add asdf plugin install instructions to README
- [x] 1.3 Create `Makefile` with targets: `build`, `test`, `lint`, `fmt`, `generate`, `install`, `clean`, `release-check`
- [x] 1.4 Create `CLAUDE.md` documenting all project conventions: commit style, no AI co-authorship, package structure, DI approach, error handling, logging, linting, testing conventions, make targets, asdf usage, and no emojis rule
- [x] 1.5 Create `.golangci.yml` using the maratori golden config from `https://github.com/maratori/golangci-lint-config`
- [x] 1.6 Create directory structure: `cmd/tokenpile/`, `internal/domain/`, `internal/store/`, `internal/provider/`, `internal/pricing/`, `internal/export/`, `internal/skill/`, `internal/tui/`, `internal/config/`
- [x] 1.7 Add all external dependencies via `go get`: `urfave/cli/v2`, `charmbracelet/bubbletea`, `charmbracelet/lipgloss`, `NimbleMarkets/ntcharts`, `modernc.org/sqlite`, `zalando/go-keyring`, `google/go-github/v68`, `stretchr/testify`, `vektra/mockery`

## 2. Domain Types

- [x] 2.1 Define `UsageEntry` struct in `internal/domain/` with all fields: ID, Repo, IssueNum, Agent, Model, TokensIn, TokensOut, SessionID, At
- [x] 2.2 Define `Session` struct in `internal/domain/` with: ID, Repo, IssueNum, StartedAt, EndedAt (*time.Time)
- [x] 2.3 Define `Issue` struct in `internal/domain/` with: Number, Repo, Title, State, URL
- [x] 2.4 Define `IssueFilter` struct in `internal/domain/` with: Repo, State, Assignee, Agent, Model, From, To
- [x] 2.5 Define `Report` struct in `internal/domain/` with: IssueNum, Repo, Rows ([]ReportRow), TotalTokensIn, TotalTokensOut, TotalCost, TotalTime
- [x] 2.6 Define `ReportRow` struct in `internal/domain/` with: Agent, Model, Calls, TokensIn, TokensOut, Cost

## 3. Configuration

- [x] 3.1 Implement `internal/config/` with XDG-compliant path resolution: config dir (`~/.config/tokenpile/`), data dir (`~/.local/share/tokenpile/`), DB path, pricing override path, identity key paths
- [x] 3.2 Write unit tests for config path resolution on macOS and Linux

## 4. Storage Layer

- [x] 4.1 Define `Store` interface in `internal/store/` with all methods: `LogUsage`, `ListIssues`, `GetReport`, `StartSession`, `EndSession`, `ListSessions`, `ListUsageOverTime`
- [x] 4.2 Implement `SQLiteStore` in `internal/store/sqlite.go` using `modernc.org/sqlite`; apply schema on first open (CREATE TABLE IF NOT EXISTS for `usage_entries` and `sessions`)
- [x] 4.3 Implement `SQLiteStore.LogUsage`: insert a `UsageEntry` with generated UUID and UTC timestamp
- [x] 4.4 Implement `SQLiteStore.StartSession` and `EndSession`: insert/update sessions table
- [x] 4.5 Implement `SQLiteStore.ListSessions`: query sessions by repo and issue_num
- [x] 4.6 Implement `SQLiteStore.ListIssues`: aggregate usage_entries grouped by (repo, issue_num) with total tokens and time
- [x] 4.7 Implement `SQLiteStore.GetReport`: aggregate usage_entries for a single issue grouped by (agent, model) with cost computed via pricing package
- [x] 4.8 Implement `SQLiteStore.ListUsageOverTime`: aggregate token usage by day or week within an optional date range and filters
- [x] 4.9 Generate `StoreMock` using mockery for use in unit tests
- [x] 4.10 Write integration tests for `SQLiteStore`: each test case uses a fresh temp DB file with `t.Cleanup` for removal; test all Store methods including edge cases (empty results, unknown model, active session)

## 5. Pricing

- [x] 5.1 Create `internal/pricing/pricing.defaults.yaml` with entries for: claude-haiku-4-5, claude-sonnet-4-6, claude-opus-4-7, gpt-4o, gpt-4o-mini, gpt-4-turbo, o1, o3-mini, gemini-1.5-pro, gemini-1.5-flash
- [x] 5.2 Embed `pricing.defaults.yaml` in the binary using `//go:embed`
- [x] 5.3 Implement `internal/pricing/` loader: parse both embedded defaults and user override file, merge (user wins), expose `ComputeCost(model string, tokensIn, tokensOut int) (float64, bool)` returning false for unknown models
- [x] 5.4 Write unit tests for pricing merge logic and cost computation
- [x] 5.5 Implement `tokenpile pricing list` command: display merged pricing config as a table, marking user-overridden entries
- [x] 5.6 Implement `tokenpile pricing set <model> --in <price> --out <price>`: add or update model in user override file

## 6. Provider Interfaces and GitHub Implementation

- [x] 6.1 Define `AuthProvider` interface in `internal/provider/`: `Login`, `Token`, `Logout`
- [x] 6.2 Define `IssueProvider` interface in `internal/provider/`: `ListIssues`, `GetIssue`
- [x] 6.3 Implement repo inference in `internal/provider/repoinfer.go`: run `git remote get-url origin`, parse HTTPS and SSH GitHub URLs into `owner/repo`; return `ErrNoRepo` sentinel if not possible
- [x] 6.4 Implement `GitHubAuthProvider` in `internal/provider/github_auth.go`: start local HTTP callback server on ephemeral port, open browser to GitHub OAuth URL, receive callback, exchange code for token, store in OS keychain via `zalando/go-keyring`, handle 2-minute timeout
- [x] 6.5 Implement OS keychain fallback for headless Linux: detect Secret Service unavailability; store AES-256-GCM encrypted token at `~/.config/tokenpile/credentials` (0600); log warning via slog
- [x] 6.6 Implement `GitHubIssueProvider` in `internal/provider/github_issues.go` using `google/go-github`; request minimum OAuth scopes (`read:user`, `repo`); implement `ListIssues` and `GetIssue`; return `ErrUnauthenticated` when token is missing or expired
- [x] 6.7 Generate `AuthProviderMock` and `IssueProviderMock` using mockery
- [x] 6.8 Write unit tests for repo inference (HTTPS, SSH, no remote, non-GitHub remote)
- [x] 6.9 Write unit tests for `GitHubIssueProvider` using mock HTTP server

## 7. Auth Commands

- [x] 7.1 Implement Ed25519 keypair generation on first run in `internal/config/identity.go`: check for `identity.key`, generate if absent, write private key (0600) and public key (0644), log to slog
- [x] 7.2 Implement `tokenpile auth login --provider <name>` command: delegate to the registered `AuthProvider` for the named provider; print success message
- [x] 7.3 Implement `tokenpile auth logout --provider <name>` command: call `AuthProvider.Logout`, remove token from keychain
- [x] 7.4 Implement `tokenpile auth status` command: check token existence for all providers, print authenticated username or "Not logged in"
- [x] 7.5 Write unit tests for auth commands using `AuthProviderMock`

## 8. Usage Tracking and Sessions

- [x] 8.1 Implement `tokenpile log` command with required flags (`--issue`, `--agent`, `--model`, `--tokens-in`, `--tokens-out`) and optional `--repo` (infer if absent, fail with clear error if inference fails)
- [x] 8.2 Implement session lifecycle in `tokenpile log`: before inserting entry, close any sessions idle >30 minutes for `(repo, issue_num)`, then start or reuse active session
- [x] 8.3 Implement `tokenpile report --issue <num> [--repo <owner/repo>]` command: print per-(agent, model) breakdown with tokens, cost (from pricing), and total wall-clock time
- [x] 8.4 Write unit tests for `log` command using `StoreMock`: test required flag validation, repo inference, session lifecycle branching
- [x] 8.5 Write integration tests for the full log + session lifecycle using a real SQLiteStore

## 9. Export and Signing

- [x] 9.1 Create `schema/export.schema.json` (JSON Schema draft 2020-12) covering the full export document structure
- [x] 9.2 Embed `schema/export.schema.json` in the binary using `//go:embed`
- [x] 9.3 Implement canonical JSON serialization in `internal/export/` (RFC 8785: deterministic key ordering)
- [x] 9.4 Implement Ed25519 sign and verify functions in `internal/export/signing.go` using `crypto/ed25519` from stdlib
- [x] 9.5 Implement `tokenpile export` command: apply scope and filter flags, query Store, compute costs, build export document, sign entries, write to stdout or `--output` file
- [x] 9.6 Implement `tokenpile export verify --file <path>`: parse file, validate against embedded JSON Schema, verify Ed25519 signature, print result
- [x] 9.7 Write unit tests for canonical JSON, signing, and verification
- [x] 9.8 Write integration tests for export + verify round-trip using a real SQLiteStore

## 10. Agent Skill Integration

- [x] 10.1 Create `internal/skill/templates/claude-code.md` with the Claude Code skill content instructing the agent to call `tokenpile log` with correct flags
- [x] 10.2 Embed all skill templates using `//go:embed`
- [x] 10.3 Implement `tokenpile skill install --agent <name>`: resolve target path for the named agent, write embedded template, print confirmation; overwrite with warning if file exists
- [x] 10.4 Implement `tokenpile skill list`: print all supported agents with installed/not-installed status
- [x] 10.5 Write unit tests for skill install path resolution and template writing

## 11. TUI

- [x] 11.1 Define Bubble Tea app structure in `internal/tui/`: top-level model with active view state, message types, and key bindings map
- [x] 11.2 Implement issue list view: scrollable list of tracked issues with number, repo, total cost, total tokens, total time; empty state message; Enter to drill down; `c` to open global chart
- [x] 11.3 Implement issue detail view: per-(agent, model) table with calls, tokens in/out, cost; total row; agent and model filter inputs; `c` to open issue chart; Esc to return to list
- [x] 11.4 Implement chart view: ASCII bar chart of token usage over time; day/week granularity toggle; cost and token totals below chart; Esc to return
- [x] 11.5 Implement `?` help overlay listing all keyboard shortcuts
- [x] 11.6 Implement auth error state in TUI: shown when no valid token, displays auth instructions
- [x] 11.7 Wire TUI to Store and IssueProvider via constructor injection
- [x] 11.8 Write unit tests for TUI model update logic (key handling, view transitions) without rendering

## 12. CLI Assembly and Main

- [x] 12.1 Implement `cmd/tokenpile/main.go` as the composition root: read config, open SQLiteStore, construct pricing loader, construct GitHubAuthProvider and GitHubIssueProvider, inject into all CLI commands and TUI
- [x] 12.2 Implement global flags: `--log-level` (debug/info/warn/error, default info) and `--log-format` (text/json, default text); initialize slog handler before any other setup
- [x] 12.3 Wire all commands into the urfave/cli app: `log`, `report`, `auth login/logout/status`, `pricing list/set`, `export`, `export verify`, `skill install/list`; default action (no args) launches TUI
- [x] 12.4 Ensure `TOKENPILE_LOG_LEVEL` and `TOKENPILE_LOG_FORMAT` environment variables override the corresponding flags
- [x] 12.5 Write smoke tests for CLI entry points using subprocess invocation

## 13. CI/CD and Release

- [x] 13.1 Create `.github/workflows/ci.yml`: trigger on push and PR to any branch; steps: checkout, setup-go, golangci-lint action, gofmt check (`gofmt -l .` fails if output non-empty), `go test -race ./...`
- [x] 13.2 Create `.goreleaser.yaml`: build targets darwin/amd64+arm64 and linux/amd64+arm64, CGO_ENABLED=0, tar.gz archives, SHA256 checksums, Homebrew tap config pointing to `cdimonaco/homebrew-tokenpile`
- [x] 13.3 Create `.github/workflows/release.yml`: trigger on `v*.*.*` tags; run goreleaser with `GITHUB_TOKEN` and `HOMEBREW_TAP_GITHUB_TOKEN` secrets
- [ ] 13.4 Create the `cdimonaco/homebrew-tokenpile` GitHub repository with an initial placeholder formula
- [x] 13.5 Validate goreleaser config locally with `goreleaser check`
