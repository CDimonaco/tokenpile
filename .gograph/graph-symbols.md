# Symbols & Packages

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-08-01 13:38:20 UTC  

---

## Important Files (top 20 by symbol+call density)

| File | Symbols | Calls |
|------|---------|-------|
| `internal/store/sqlite_test.go` | 36 | 384 |
| `cmd/tokenpile/integration_test.go` | 32 | 304 |
| `internal/store/sqlite.go` | 37 | 282 |
| `internal/tui/tui.go` | 53 | 258 |
| `internal/tui/tui_test.go` | 41 | 228 |
| `internal/skill/skill_test.go` | 21 | 147 |
| `cmd/tokenpile/cmd_hook_test.go` | 15 | 151 |
| `internal/export/export_test.go` | 21 | 145 |
| `cmd/tokenpile/cmd_auth_test.go` | 23 | 131 |
| `internal/pricing/pricing_test.go` | 17 | 122 |
| `internal/provider/github_auth.go` | 17 | 98 |
| `cmd/tokenpile/cmd_reset_test.go` | 8 | 98 |
| `internal/capture/capture_test.go` | 15 | 90 |
| `internal/provider/github_issues_test.go` | 8 | 96 |
| `internal/provider/ghcli_auth_test.go` | 14 | 86 |
| `internal/skill/hooks_test.go` | 9 | 86 |
| `internal/skill/skill.go` | 21 | 70 |
| `cmd/tokenpile/smoke_test.go` | 8 | 81 |
| `cmd/tokenpile/cmd_report.go` | 6 | 76 |
| `cmd/tokenpile/cmd_log_test.go` | 7 | 74 |

## Important Symbols (top 30 by outgoing calls)

| Symbol | Kind | File | Line | Calls out |
|--------|------|------|------|-----------|
| `main` | function | `cmd/tokenpile/main.go` | 24 | 46 |
| `(*GitHubAuthProvider).Login` | method | `internal/provider/github_auth.go` | 55 | 43 |
| `newResetFixture` | function | `cmd/tokenpile/cmd_reset_test.go` | 37 | 42 |
| `TestTUI_DetailView_SessionsTab_ShowsNoteAndTags` | function | `internal/tui/tui_test.go` | 392 | 36 |
| `runLog` | function | `cmd/tokenpile/cmd_log.go` | 82 | 32 |
| `TestSmoke_ExportVerify` | function | `cmd/tokenpile/smoke_test.go` | 70 | 31 |
| `(Model).viewIssueList` | method | `internal/tui/tui.go` | 434 | 30 |
| `reportCommand` | function | `cmd/tokenpile/cmd_report.go` | 19 | 29 |
| `(*SQLiteStore).ListUsageOverTime` | method | `internal/store/sqlite.go` | 700 | 27 |
| `unattributedCommand` | function | `cmd/tokenpile/cmd_bind.go` | 84 | 26 |
| `(*SQLiteStore).ListIssues` | method | `internal/store/sqlite.go` | 556 | 26 |
| `TestLog_ClosesIdleSession_StartsNew` | function | `cmd/tokenpile/cmd_log_test.go` | 126 | 25 |
| `Build` | function | `internal/export/export.go` | 82 | 25 |
| `TestSQLiteStore_AssignIssue_AndReverse` | function | `internal/store/sqlite_test.go` | 630 | 25 |
| `(Model).viewIssueDetail` | method | `internal/tui/tui.go` | 493 | 25 |
| `runExport` | function | `cmd/tokenpile/cmd_export.go` | 230 | 24 |
| `(*SQLiteStore).ListEntries` | method | `internal/store/sqlite.go` | 218 | 24 |
| `(Model).renderFooter` | method | `internal/tui/tui.go` | 391 | 24 |
| `TestGitHubIssueProvider_ListIssues_Paginated` | function | `internal/provider/github_issues_test.go` | 49 | 23 |
| `TestSQLiteStore_ListTrackedIssueRefs` | function | `internal/store/sqlite_test.go` | 318 | 23 |
| `TestSQLiteStore_ListUsageOverTime_WeekGranularity` | function | `internal/store/sqlite_test.go` | 246 | 23 |
| `runVerify` | function | `cmd/tokenpile/cmd_export.go` | 83 | 22 |
| `(Model).renderSessionsTab` | method | `internal/tui/tui.go` | 591 | 22 |
| `TestTUI_DetailView_SummaryTab_ShowsBudgetBar` | function | `internal/tui/tui_test.go` | 433 | 22 |
| `TestLog_NoActiveSession_StartsNew` | function | `cmd/tokenpile/cmd_log_test.go` | 51 | 21 |
| `TestReconcile_UnattributableTurnDoesNotStallTheSpool` | function | `cmd/tokenpile/cmd_hook_test.go` | 224 | 21 |
| `writeResetBackup` | function | `cmd/tokenpile/cmd_reset.go` | 152 | 21 |
| `(*SQLiteStore).ListUnattributed` | method | `internal/store/sqlite.go` | 1010 | 21 |
| `budgetCommands` | function | `cmd/tokenpile/cmd_budget.go` | 13 | 20 |
| `TestIntegration_Log_IdleSessionClosed_ByResolveSession` | function | `cmd/tokenpile/integration_test.go` | 694 | 19 |

## Packages

| Package | Dir | Files | Symbols |
|---------|-----|-------|---------|
| `main` | `cmd/tokenpile` | 18 | 149 |
| `attribution` | `internal/attribution` | 2 | 12 |
| `capture` | `internal/capture` | 5 | 20 |
| `config` | `internal/config` | 3 | 8 |
| `export` | `internal/export` | 4 | 15 |
| `pricing` | `internal/pricing` | 2 | 13 |
| `provider` | `internal/provider` | 13 | 65 |
| `schema` | `internal/schema` | 1 | 1 |
| `skill` | `internal/skill` | 4 | 36 |
| `store` | `internal/store` | 3 | 38 |
| `tui` | `internal/tui` | 2 | 94 |
| `usage` | `internal/usage` | 1 | 22 |

