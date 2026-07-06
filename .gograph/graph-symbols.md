# Symbols & Packages

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-07-06 20:45:13 UTC  

---

## Important Files (top 20 by symbol+call density)

| File | Symbols | Calls |
|------|---------|-------|
| `cmd/tokenpile/integration_test.go` | 32 | 868 |
| `internal/store/sqlite.go` | 33 | 834 |
| `internal/store/sqlite_test.go` | 30 | 758 |
| `internal/tui/tui.go` | 53 | 547 |
| `internal/tui/tui_test.go` | 40 | 511 |
| `internal/skill/skill_test.go` | 26 | 483 |
| `internal/export/export_test.go` | 22 | 425 |
| `cmd/tokenpile/cmd_reset_test.go` | 8 | 272 |
| `internal/provider/github_auth.go` | 15 | 228 |
| `cmd/tokenpile/smoke_test.go` | 8 | 207 |
| `cmd/tokenpile/cmd_reset.go` | 7 | 206 |
| `cmd/tokenpile/cmd_export.go` | 5 | 206 |
| `internal/export/export.go` | 14 | 189 |
| `internal/skill/skill.go` | 22 | 170 |
| `cmd/tokenpile/cmd_log.go` | 5 | 164 |
| `cmd/tokenpile/cmd_log_test.go` | 6 | 154 |
| `internal/export/export_integration_test.go` | 6 | 150 |
| `cmd/tokenpile/cmd_report.go` | 2 | 143 |
| `cmd/tokenpile/main.go` | 5 | 137 |
| `internal/provider/github_issues_test.go` | 5 | 135 |

## Important Symbols (top 30 by outgoing calls)

| Symbol | Kind | File | Line | Calls out |
|--------|------|------|------|-----------|
| `main` | function | `cmd/tokenpile/main.go` | 24 | 118 |
| `(*SQLiteStore).ListIssues` | method | `internal/store/sqlite.go` | 524 | 108 |
| `(*SQLiteStore).ListUsageOverTime` | method | `internal/store/sqlite.go` | 661 | 98 |
| `newResetFixture` | function | `cmd/tokenpile/cmd_reset_test.go` | 37 | 96 |
| `reportCommand` | function | `cmd/tokenpile/cmd_report.go` | 15 | 95 |
| `(*SQLiteStore).ListEntries` | method | `internal/store/sqlite.go` | 206 | 94 |
| `(*GitHubAuthProvider).Login` | method | `internal/provider/github_auth.go` | 55 | 93 |
| `TestSmoke_ExportVerify` | function | `cmd/tokenpile/smoke_test.go` | 70 | 90 |
| `runLog` | function | `cmd/tokenpile/cmd_log.go` | 73 | 89 |
| `Build` | function | `internal/export/export.go` | 84 | 79 |
| `TestTUI_DetailView_SessionsTab_ShowsNoteAndTags` | function | `internal/tui/tui_test.go` | 392 | 75 |
| `runExport` | function | `cmd/tokenpile/cmd_export.go` | 226 | 73 |
| `(Model).viewIssueList` | method | `internal/tui/tui.go` | 434 | 68 |
| `(Model).loadIssues` | method | `internal/tui/tui.go` | 771 | 62 |
| `runReset` | function | `cmd/tokenpile/cmd_reset.go` | 55 | 58 |
| `runVerify` | function | `cmd/tokenpile/cmd_export.go` | 83 | 58 |
| `scanSessions` | function | `internal/store/sqlite.go` | 393 | 58 |
| `writeResetBackup` | function | `cmd/tokenpile/cmd_reset.go` | 149 | 57 |
| `TestIntegration_ExportVerify_TamperedSessionFails` | function | `cmd/tokenpile/integration_test.go` | 621 | 56 |
| `TestIntegration_Export_RepoFilter_ScopesSessionsAndBudgets` | function | `cmd/tokenpile/integration_test.go` | 564 | 56 |
| `(Model).viewIssueDetail` | method | `internal/tui/tui.go` | 493 | 55 |
| `TestIntegration_Reset_YesResetsAndBackupVerifies` | function | `cmd/tokenpile/cmd_reset_test.go` | 118 | 53 |
| `TestGenerateFixtures` | function | `internal/export/gen_fixture_internal_test.go` | 14 | 53 |
| `(*SQLiteStore).UpdateSessionAnnotations` | method | `internal/store/sqlite.go` | 459 | 53 |
| `resolveSession` | function | `cmd/tokenpile/cmd_log.go` | 170 | 52 |
| `(*SQLiteStore).GetReport` | method | `internal/store/sqlite.go` | 618 | 52 |
| `TestIntegration_Export_Unfiltered_IncludesAllSessionsAndBudgets` | function | `cmd/tokenpile/integration_test.go` | 538 | 51 |
| `marshalCanonical` | function | `internal/export/export.go` | 238 | 51 |
| `TestSQLiteStore_ListUsageOverTime_WeekGranularity` | function | `internal/store/sqlite_test.go` | 196 | 51 |
| `TestIntegration_Log_IdleSessionClosed_ByResolveSession` | function | `cmd/tokenpile/integration_test.go` | 703 | 50 |

## Packages

| Package | Dir | Files | Symbols |
|---------|-----|-------|---------|
| `main` | `cmd/tokenpile` | 15 | 94 |
| `config` | `internal/config` | 3 | 8 |
| `export` | `internal/export` | 5 | 16 |
| `pricing` | `internal/pricing` | 2 | 9 |
| `provider` | `internal/provider` | 7 | 35 |
| `schema` | `internal/schema` | 1 | 1 |
| `skill` | `internal/skill` | 2 | 22 |
| `store` | `internal/store` | 3 | 34 |
| `tui` | `internal/tui` | 2 | 93 |
| `usage` | `internal/usage` | 1 | 14 |

