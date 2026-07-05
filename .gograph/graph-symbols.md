# Symbols & Packages

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-07-05 14:37:41 UTC  

---

## Important Files (top 20 by symbol+call density)

| File | Symbols | Calls |
|------|---------|-------|
| `internal/store/sqlite.go` | 28 | 758 |
| `internal/store/sqlite_test.go` | 27 | 689 |
| `internal/tui/tui.go` | 53 | 547 |
| `cmd/tokenpile/integration_test.go` | 23 | 562 |
| `internal/tui/tui_test.go` | 40 | 511 |
| `internal/skill/skill_test.go` | 20 | 349 |
| `internal/export/export_test.go` | 17 | 318 |
| `internal/provider/github_auth.go` | 14 | 208 |
| `internal/skill/skill.go` | 18 | 168 |
| `internal/export/export.go` | 11 | 169 |
| `cmd/tokenpile/cmd_log.go` | 5 | 161 |
| `cmd/tokenpile/cmd_log_test.go` | 6 | 154 |
| `internal/export/export_integration_test.go` | 6 | 141 |
| `cmd/tokenpile/cmd_report.go` | 2 | 143 |
| `cmd/tokenpile/main.go` | 5 | 126 |
| `internal/pricing/pricing_test.go` | 7 | 121 |
| `cmd/tokenpile/cmd_export.go` | 2 | 120 |
| `cmd/tokenpile/smoke_test.go` | 7 | 112 |
| `internal/config/paths_test.go` | 4 | 112 |
| `internal/provider/github_issues.go` | 6 | 104 |

## Important Symbols (top 30 by outgoing calls)

| Symbol | Kind | File | Line | Calls out |
|--------|------|------|------|-----------|
| `(*SQLiteStore).ListIssues` | method | `internal/store/sqlite.go` | 505 | 110 |
| `main` | function | `cmd/tokenpile/main.go` | 24 | 107 |
| `(*SQLiteStore).ListUsageOverTime` | method | `internal/store/sqlite.go` | 644 | 98 |
| `reportCommand` | function | `cmd/tokenpile/cmd_report.go` | 15 | 95 |
| `(*GitHubAuthProvider).Login` | method | `internal/provider/github_auth.go` | 53 | 94 |
| `(*SQLiteStore).ListEntries` | method | `internal/store/sqlite.go` | 206 | 94 |
| `runExport` | function | `cmd/tokenpile/cmd_export.go` | 97 | 89 |
| `runLog` | function | `cmd/tokenpile/cmd_log.go` | 73 | 89 |
| `Build` | function | `internal/export/export.go` | 69 | 75 |
| `TestTUI_DetailView_SessionsTab_ShowsNoteAndTags` | function | `internal/tui/tui_test.go` | 392 | 75 |
| `(*SQLiteStore).ListSessions` | method | `internal/store/sqlite.go` | 363 | 68 |
| `(Model).viewIssueList` | method | `internal/tui/tui.go` | 434 | 68 |
| `(Model).loadIssues` | method | `internal/tui/tui.go` | 771 | 62 |
| `(Model).viewIssueDetail` | method | `internal/tui/tui.go` | 493 | 55 |
| `(*SQLiteStore).UpdateSessionAnnotations` | method | `internal/store/sqlite.go` | 440 | 53 |
| `resolveSession` | function | `cmd/tokenpile/cmd_log.go` | 170 | 52 |
| `(*SQLiteStore).GetReport` | method | `internal/store/sqlite.go` | 601 | 52 |
| `marshalCanonical` | function | `internal/export/export.go` | 186 | 51 |
| `TestInstall_Codex_AppendsToExistingFile` | function | `internal/skill/skill_test.go` | 95 | 51 |
| `installShared` | function | `internal/skill/skill.go` | 124 | 51 |
| `TestSQLiteStore_ListUsageOverTime_WeekGranularity` | function | `internal/store/sqlite_test.go` | 196 | 51 |
| `TestIntegration_Log_IdleSessionClosed_ByResolveSession` | function | `cmd/tokenpile/integration_test.go` | 537 | 50 |
| `budgetCommands` | function | `cmd/tokenpile/cmd_budget.go` | 13 | 49 |
| `printSessionsReport` | function | `cmd/tokenpile/cmd_report.go` | 115 | 48 |
| `(*GitHubIssueProvider).ListIssues` | method | `internal/provider/github_issues.go` | 52 | 48 |
| `TestLog_ClosesIdleSession_StartsNew` | function | `cmd/tokenpile/cmd_log_test.go` | 126 | 45 |
| `TestSQLiteStore_UpdateSessionAnnotations_TagsUnion` | function | `internal/store/sqlite_test.go` | 330 | 45 |
| `TestResolve_UsesEnvOverride` | function | `internal/config/paths_test.go` | 14 | 44 |
| `TestSQLiteStore_ListTrackedIssueRefs` | function | `internal/store/sqlite_test.go` | 266 | 44 |
| `pricingCommands` | function | `cmd/tokenpile/cmd_pricing.go` | 13 | 43 |

## Packages

| Package | Dir | Files | Symbols |
|---------|-----|-------|---------|
| `main` | `cmd/tokenpile` | 13 | 67 |
| `config` | `internal/config` | 3 | 8 |
| `export` | `internal/export` | 3 | 11 |
| `pricing` | `internal/pricing` | 2 | 9 |
| `provider` | `internal/provider` | 6 | 31 |
| `schema` | `internal/schema` | 1 | 1 |
| `skill` | `internal/skill` | 2 | 18 |
| `store` | `internal/store` | 3 | 29 |
| `tui` | `internal/tui` | 2 | 93 |
| `usage` | `internal/usage` | 1 | 13 |

