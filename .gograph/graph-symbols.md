# Symbols & Packages

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-07-05 15:02:06 UTC  

---

## Important Files (top 20 by symbol+call density)

| File | Symbols | Calls |
|------|---------|-------|
| `internal/store/sqlite.go` | 30 | 800 |
| `cmd/tokenpile/integration_test.go` | 31 | 793 |
| `internal/store/sqlite_test.go` | 27 | 689 |
| `internal/tui/tui.go` | 53 | 547 |
| `internal/tui/tui_test.go` | 40 | 511 |
| `internal/export/export_test.go` | 22 | 425 |
| `internal/skill/skill_test.go` | 20 | 349 |
| `internal/provider/github_auth.go` | 15 | 228 |
| `internal/export/export.go` | 14 | 189 |
| `cmd/tokenpile/cmd_export.go` | 4 | 194 |
| `internal/skill/skill.go` | 18 | 168 |
| `cmd/tokenpile/cmd_log.go` | 5 | 164 |
| `cmd/tokenpile/cmd_log_test.go` | 6 | 154 |
| `internal/export/export_integration_test.go` | 6 | 150 |
| `cmd/tokenpile/cmd_report.go` | 2 | 143 |
| `internal/provider/github_issues_test.go` | 5 | 135 |
| `cmd/tokenpile/main.go` | 5 | 126 |
| `internal/pricing/pricing_test.go` | 7 | 121 |
| `cmd/tokenpile/smoke_test.go` | 7 | 112 |
| `internal/config/paths_test.go` | 4 | 112 |

## Important Symbols (top 30 by outgoing calls)

| Symbol | Kind | File | Line | Calls out |
|--------|------|------|------|-----------|
| `(*SQLiteStore).ListIssues` | method | `internal/store/sqlite.go` | 505 | 108 |
| `main` | function | `cmd/tokenpile/main.go` | 24 | 107 |
| `(*SQLiteStore).ListUsageOverTime` | method | `internal/store/sqlite.go` | 642 | 98 |
| `reportCommand` | function | `cmd/tokenpile/cmd_report.go` | 15 | 95 |
| `(*SQLiteStore).ListEntries` | method | `internal/store/sqlite.go` | 206 | 94 |
| `(*GitHubAuthProvider).Login` | method | `internal/provider/github_auth.go` | 55 | 93 |
| `runExport` | function | `cmd/tokenpile/cmd_export.go` | 175 | 89 |
| `runLog` | function | `cmd/tokenpile/cmd_log.go` | 73 | 89 |
| `Build` | function | `internal/export/export.go` | 84 | 79 |
| `TestTUI_DetailView_SessionsTab_ShowsNoteAndTags` | function | `internal/tui/tui_test.go` | 392 | 75 |
| `(*SQLiteStore).ListSessions` | method | `internal/store/sqlite.go` | 363 | 68 |
| `(Model).viewIssueList` | method | `internal/tui/tui.go` | 434 | 68 |
| `(Model).loadIssues` | method | `internal/tui/tui.go` | 771 | 62 |
| `runVerify` | function | `cmd/tokenpile/cmd_export.go` | 82 | 58 |
| `TestIntegration_ExportVerify_TamperedSessionFails` | function | `cmd/tokenpile/integration_test.go` | 587 | 56 |
| `(Model).viewIssueDetail` | method | `internal/tui/tui.go` | 493 | 55 |
| `TestGenerateFixtures` | function | `internal/export/gen_fixture_internal_test.go` | 14 | 53 |
| `(*SQLiteStore).UpdateSessionAnnotations` | method | `internal/store/sqlite.go` | 440 | 53 |
| `resolveSession` | function | `cmd/tokenpile/cmd_log.go` | 170 | 52 |
| `(*SQLiteStore).GetReport` | method | `internal/store/sqlite.go` | 599 | 52 |
| `marshalCanonical` | function | `internal/export/export.go` | 238 | 51 |
| `TestInstall_Codex_AppendsToExistingFile` | function | `internal/skill/skill_test.go` | 95 | 51 |
| `installShared` | function | `internal/skill/skill.go` | 124 | 51 |
| `TestSQLiteStore_ListUsageOverTime_WeekGranularity` | function | `internal/store/sqlite_test.go` | 196 | 51 |
| `TestIntegration_Log_IdleSessionClosed_ByResolveSession` | function | `cmd/tokenpile/integration_test.go` | 669 | 50 |
| `budgetCommands` | function | `cmd/tokenpile/cmd_budget.go` | 13 | 49 |
| `printSessionsReport` | function | `cmd/tokenpile/cmd_report.go` | 115 | 48 |
| `TestLog_ClosesIdleSession_StartsNew` | function | `cmd/tokenpile/cmd_log_test.go` | 126 | 45 |
| `TestSQLiteStore_UpdateSessionAnnotations_TagsUnion` | function | `internal/store/sqlite_test.go` | 330 | 45 |
| `TestResolve_UsesEnvOverride` | function | `internal/config/paths_test.go` | 14 | 44 |

## Packages

| Package | Dir | Files | Symbols |
|---------|-----|-------|---------|
| `main` | `cmd/tokenpile` | 13 | 77 |
| `config` | `internal/config` | 3 | 8 |
| `export` | `internal/export` | 5 | 16 |
| `pricing` | `internal/pricing` | 2 | 9 |
| `provider` | `internal/provider` | 7 | 35 |
| `schema` | `internal/schema` | 1 | 1 |
| `skill` | `internal/skill` | 2 | 18 |
| `store` | `internal/store` | 3 | 31 |
| `tui` | `internal/tui` | 2 | 93 |
| `usage` | `internal/usage` | 1 | 13 |

