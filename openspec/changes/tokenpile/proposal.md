## Why

Developers using LLM agents (Claude Code, OpenCode, Cursor, Codex, etc.) have no visibility into how many tokens they spend, how much that costs, or how much time they invest per GitHub issue. tokenpile fills that gap with a local-first, agent-agnostic CLI and TUI that tracks usage at the issue level.

## What Changes

This is a greenfield project. All capabilities are new.

- A CLI (`tokenpile`) that agents can call to log token usage against a GitHub issue
- A TUI launched by `tokenpile` (no args) to browse issues, view breakdowns, and explore charts
- GitHub OAuth via a local callback server, behind an extensible provider interface
- Local SQLite storage behind a swappable `Store` interface
- Session tracking for wall-clock time per issue (30-minute idle auto-close)
- Two-layer pricing configuration (embedded defaults + user overrides) for cost computation
- Ed25519 signing of JSON exports for tamper-evidence
- Agent skill installation via `tokenpile skill install --agent <name>`
- goreleaser-based releases: macOS (amd64/arm64), Linux (amd64/arm64), Homebrew tap
- GitHub Actions CI: lint, format, and test on every push; release on tag
- `CLAUDE.md` documenting all project conventions for agents working on this repo

## Capabilities

### New Capabilities

- `auth`: OAuth 2.0 local-callback flow behind an `AuthProvider` interface; token storage in OS keychain via `zalando/go-keyring`
- `issue-provider`: `IssueProvider` interface with GitHub REST implementation; minimum required OAuth scopes
- `usage-tracking`: `tokenpile log` command; `UsageEntry` persistence; agent + model both required; repo inferred from git remote when not explicit
- `sessions`: implicit session lifecycle tied to issue; 30-minute idle auto-close; wall-clock time aggregation
- `pricing`: embedded default pricing YAML + user override file; cost computed at report time, never stored
- `tui`: Bubble Tea TUI with three views — issue list, issue detail (per-agent/model breakdown), usage charts (day/week granularity, filterable by agent and model)
- `export`: `tokenpile export` with scope and filter flags; self-contained signed JSON with embedded Ed25519 public key and signature; JSON Schema for validation; `tokenpile export verify`
- `agent-skill`: `tokenpile skill install --agent <name>` embeds and writes agent-specific skill files (e.g., Claude Code skill at `~/.claude/skills/tokenpile.md`)
- `ci-release`: GitHub Actions workflows for CI and release; goreleaser config; Homebrew tap at `cdimonaco/homebrew-tokenpile`

### Modified Capabilities

## Impact

- New Go module: `github.com/cdimonaco/tokenpile`
- External dependencies: `urfave/cli`, `charmbracelet/bubbletea`, `charmbracelet/lipgloss`, `NimbleMarkets/ntcharts`, `modernc.org/sqlite`, `zalando/go-keyring`, `google/go-github`, `testify`
- Local state: `~/.config/tokenpile/` (pricing override, Ed25519 keypair), OS keychain (OAuth tokens), SQLite DB at `~/.local/share/tokenpile/tokenpile.db`
- External services: GitHub OAuth app (read:user + repo scopes)
- CI/CD: GitHub Actions, goreleaser, separate `cdimonaco/homebrew-tokenpile` repository
