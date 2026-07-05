# GoGraph Report Index

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-07-05 15:02:06 UTC  

---

## 1. Summary

| Metric | Count |
|--------|-------|
| Packages | 10 |
| Files | 39 |
| Symbols | 395 |
| Functions | 246 |
| Methods | 57 |
| Structs | 36 |
| Interfaces | 4 |
| Call edges | 6913 |
| Env var reads | 7 |

## 2. Structural Index

> **🚨 CRITICAL WARNING FOR AI AGENTS:** DO NOT READ `graph.json` DIRECTLY. It is a massive database file that will exhaust your context window and crash your session. Use the `gograph` CLI commands (e.g., `gograph query`, `gograph focus`) to extract targeted slices of data instead.

To save token context, the full graph report has been split into targeted files. Read only what you need:

| Category | File | Description |
|----------|------|-------------|
| **Symbols** | [`graph-symbols.md`](graph-symbols.md) | Top files, heavily called symbols, and package layouts |
| **Deps** | [`graph-deps.md`](graph-deps.md) | `go.mod` tech stack and package import relationships |
| **Config** | [`graph-config.md`](graph-config.md) | Every `os.Getenv` and configuration read across the repo |
| **Concurrency** | [`graph-concurrency.md`](graph-concurrency.md) | Goroutines, channels, mutexes, and WaitGroups |
| **Routes** | [`graph-routes.md`](graph-routes.md) | HTTP REST API routes and handlers |
| **SQL** | [`graph-sql.md`](graph-sql.md) | Raw database queries mapped to functions |
| **Errors** | [`graph-errors.md`](graph-errors.md) | Custom errors and panics mapped to origin lines |
| **Tests** | [`graph-tests.md`](graph-tests.md) | Which test functions exercise which production symbols |

## 3. Likely Entry Points

- `cmd/tokenpile/cmd_auth.go` (package `main`)
- `cmd/tokenpile/cmd_auth_test.go` (package `main`)
- `cmd/tokenpile/cmd_budget.go` (package `main`)
- `cmd/tokenpile/cmd_budget_test.go` (package `main`)
- `cmd/tokenpile/cmd_export.go` (package `main`)
- `cmd/tokenpile/cmd_log.go` (package `main`)
- `cmd/tokenpile/cmd_log_test.go` (package `main`)
- `cmd/tokenpile/cmd_pricing.go` (package `main`)
- `cmd/tokenpile/cmd_report.go` (package `main`)
- `cmd/tokenpile/cmd_skill.go` (package `main`)
- `cmd/tokenpile/integration_test.go` (package `main`)
- `cmd/tokenpile/main.go` (package `main`)
- `cmd/tokenpile/smoke_test.go` (package `main`)

## AI Assistant / Coding Agent Usage

Add the following to your AI assistant's system prompt, project instructions, or context file:

> Before answering architecture, dependency, or 'where is X?' questions about this
> repository, read `.gograph/GRAPH_REPORT.md` first. Use it as the repo map before
> searching raw files. Use `gograph query` and `gograph callers` for symbol lookup.
