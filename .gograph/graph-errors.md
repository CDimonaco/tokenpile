# Errors & Panics

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-07-05 14:42:54 UTC  

---

| Message | Function | File | Line |
|---------|----------|------|------|
| `login: %w` | `authCommands` | `cmd/tokenpile/cmd_auth.go` | 29 |
| `logout: %w` | `authCommands` | `cmd/tokenpile/cmd_auth.go` | 50 |
| `oauth failed` | `TestAuthLogin_Failure` | `cmd/tokenpile/cmd_auth_test.go` | 44 |
| `not found` | `TestAuthStatus_NotLoggedIn` | `cmd/tokenpile/cmd_auth_test.go` | 76 |
| `cannot infer repo: pass --repo owner/repo or run from inside a GitHub repository` | `budgetCommands` | `cmd/tokenpile/cmd_budget.go` | 44 |
| `infer repo: %w` | `budgetCommands` | `cmd/tokenpile/cmd_budget.go` | 49 |
| `--amount must be greater than zero` | `budgetCommands` | `cmd/tokenpile/cmd_budget.go` | 54 |
| `set budget: %w` | `budgetCommands` | `cmd/tokenpile/cmd_budget.go` | 60 |
| `cannot infer repo: pass --repo owner/repo or run from inside a GitHub repository` | `budgetCommands` | `cmd/tokenpile/cmd_budget.go` | 88 |
| `infer repo: %w` | `budgetCommands` | `cmd/tokenpile/cmd_budget.go` | 93 |
| `unset budget: %w` | `budgetCommands` | `cmd/tokenpile/cmd_budget.go` | 99 |
| `read file: %w` | `exportCommands` | `cmd/tokenpile/cmd_export.go` | 71 |
| `parse export: %w` | `exportCommands` | `cmd/tokenpile/cmd_export.go` | 76 |
| `verification failed: %w` | `exportCommands` | `cmd/tokenpile/cmd_export.go` | 82 |
| `parse --from: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 109 |
| `parse --to: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 118 |
| `list entries: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 126 |
| `list sessions: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 133 |
| `get budget: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 141 |
| `build export: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 151 |
| `marshal export: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 156 |
| `write output: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 161 |
| `cannot infer repo: pass --repo owner/repo or run from inside a GitHub repository` | `runLog` | `cmd/tokenpile/cmd_log.go` | 77 |
| `infer repo: %w` | `runLog` | `cmd/tokenpile/cmd_log.go` | 82 |
| `--tokens-in must be zero or greater` | `runLog` | `cmd/tokenpile/cmd_log.go` | 93 |
| `--tokens-out must be zero or greater` | `runLog` | `cmd/tokenpile/cmd_log.go` | 97 |
| `issue #%d not found in %s` | `runLog` | `cmd/tokenpile/cmd_log.go` | 103 |
| `GitHub authentication required to validate issues: run tokenpile auth login` | `runLog` | `cmd/tokenpile/cmd_log.go` | 107 |
| `validate issue: %w` | `runLog` | `cmd/tokenpile/cmd_log.go` | 110 |
| `resolve session: %w` | `runLog` | `cmd/tokenpile/cmd_log.go` | 124 |
| `log usage: %w` | `runLog` | `cmd/tokenpile/cmd_log.go` | 140 |
| `list sessions: %w` | `resolveSession` | `cmd/tokenpile/cmd_log.go` | 173 |
| `end idle session: %w` | `resolveSession` | `cmd/tokenpile/cmd_log.go` | 187 |
| `start session: %w` | `resolveSession` | `cmd/tokenpile/cmd_log.go` | 204 |
| `model name is required` | `pricingCommands` | `cmd/tokenpile/cmd_pricing.go` | 66 |
| `set pricing: %w` | `pricingCommands` | `cmd/tokenpile/cmd_pricing.go` | 73 |
| `cannot infer repo: pass --repo owner/repo or run from inside a GitHub repository` | `reportCommand` | `cmd/tokenpile/cmd_report.go` | 40 |
| `infer repo: %w` | `reportCommand` | `cmd/tokenpile/cmd_report.go` | 45 |
| `get report: %w` | `reportCommand` | `cmd/tokenpile/cmd_report.go` | 57 |
| `get issue cache: %w` | `reportCommand` | `cmd/tokenpile/cmd_report.go` | 62 |
| `get budget: %w` | `reportCommand` | `cmd/tokenpile/cmd_report.go` | 101 |
| `list sessions: %w` | `printSessionsReport` | `cmd/tokenpile/cmd_report.go` | 120 |
| `install skill: %w` | `skillCommands` | `cmd/tokenpile/cmd_skill.go` | 41 |
| `` | `newTestStore` | `cmd/tokenpile/integration_test.go` | 40 |
| `tui: %w` | `main` | `cmd/tokenpile/main.go` | 112 |
| `generate ed25519 key: %w` | `generateIdentity` | `internal/config/identity.go` | 24 |
| `write identity key: %w` | `generateIdentity` | `internal/config/identity.go` | 38 |
| `write identity pub: %w` | `generateIdentity` | `internal/config/identity.go` | 42 |
| `read identity key: %w` | `loadIdentity` | `internal/config/identity.go` | 53 |
| `decode identity key PEM` | `loadIdentity` | `internal/config/identity.go` | 58 |
| `invalid identity key size: got %d, want %d` | `loadIdentity` | `internal/config/identity.go` | 62 |
| `private key is not ed25519` | `loadIdentity` | `internal/config/identity.go` | 69 |
| `canonical JSON: %w` | `Build` | `internal/export/export.go` | 95 |
| `private key is not ed25519` | `Build` | `internal/export/export.go` | 103 |
| `decode public key: %w` | `Verify` | `internal/export/export.go` | 144 |
| `invalid public key size: got %d, want %d` | `Verify` | `internal/export/export.go` | 148 |
| `decode signature: %w` | `Verify` | `internal/export/export.go` | 155 |
| `canonical JSON: %w` | `Verify` | `internal/export/export.go` | 160 |
| `signature invalid: entries have been tampered with` | `Verify` | `internal/export/export.go` | 166 |
| `parse default pricing: %w` | `NewLoader` | `internal/pricing/pricing.go` | 31 |
| `read pricing override: %w` | `NewLoader` | `internal/pricing/pricing.go` | 40 |
| `parse pricing override: %w` | `NewLoader` | `internal/pricing/pricing.go` | 46 |
| `read pricing override: %w` | `(*Loader).SetOverride` | `internal/pricing/pricing.go` | 78 |
| `parse pricing override: %w` | `(*Loader).SetOverride` | `internal/pricing/pricing.go` | 84 |
| `marshal pricing override: %w` | `(*Loader).SetOverride` | `internal/pricing/pricing.go` | 99 |
| `write pricing override: %w` | `(*Loader).SetOverride` | `internal/pricing/pricing.go` | 103 |
| `` | `TestNewLoader_DefaultsLoaded` | `internal/pricing/pricing_test.go` | 15 |
| `` | `TestComputeCost_UnknownModel` | `internal/pricing/pricing_test.go` | 72 |
| `` | `TestComputeCost_InOutSeparate` | `internal/pricing/pricing_test.go` | 81 |
| `` | `TestSetOverride_WritesAndUpdates` | `internal/pricing/pricing_test.go` | 93 |
| `start callback server: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 58 |
| `unexpected listener address type %T` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 64 |
| `oauth callback missing code` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 95 |
| `oauth callback: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 129 |
| `login timed out, please try again` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 132 |
| `exchange oauth code: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 142 |
| `store token: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 146 |
| `delete token from keychain: %w` | `(*GitHubAuthProvider).Logout` | `internal/provider/github_auth.go` | 172 |
| `remove credentials file: %w` | `(*GitHubAuthProvider).Logout` | `internal/provider/github_auth.go` | 179 |
| `create cipher: %w` | `(*GitHubAuthProvider).storeEncryptedToken` | `internal/provider/github_auth.go` | 201 |
| `create GCM: %w` | `(*GitHubAuthProvider).storeEncryptedToken` | `internal/provider/github_auth.go` | 206 |
| `generate nonce: %w` | `(*GitHubAuthProvider).storeEncryptedToken` | `internal/provider/github_auth.go` | 211 |
| `write credentials: %w` | `(*GitHubAuthProvider).storeEncryptedToken` | `internal/provider/github_auth.go` | 217 |
| `read credentials: %w` | `(*GitHubAuthProvider).loadEncryptedToken` | `internal/provider/github_auth.go` | 226 |
| `create cipher: %w` | `(*GitHubAuthProvider).loadEncryptedToken` | `internal/provider/github_auth.go` | 233 |
| `create GCM: %w` | `(*GitHubAuthProvider).loadEncryptedToken` | `internal/provider/github_auth.go` | 238 |
| `credentials file corrupted` | `(*GitHubAuthProvider).loadEncryptedToken` | `internal/provider/github_auth.go` | 243 |
| `decrypt credentials: %w` | `(*GitHubAuthProvider).loadEncryptedToken` | `internal/provider/github_auth.go` | 250 |
| `unsupported platform: %s` | `openBrowser` | `internal/provider/github_auth.go` | 282 |
| `set base URL: %w` | `(*GitHubIssueProvider).client` | `internal/provider/github_issues.go` | 45 |
| `invalid repo format %q: expected owner/repo` | `(*GitHubIssueProvider).ListIssues` | `internal/provider/github_issues.go` | 60 |
| `list issues: %w` | `(*GitHubIssueProvider).ListIssues` | `internal/provider/github_issues.go` | 80 |
| `invalid repo format %q: expected owner/repo` | `(*GitHubIssueProvider).GetIssue` | `internal/provider/github_issues.go` | 121 |
| `get issue: %w` | `(*GitHubIssueProvider).GetIssue` | `internal/provider/github_issues.go` | 133 |
| `cannot infer repo from remote %q: not a GitHub remote; pass --repo owner/repo` | `ParseRemote` | `internal/provider/repoinfer.go` | 46 |
| `%w: %s` | `Install` | `internal/skill/skill.go` | 94 |
| `cannot determine install path for agent %s` | `Install` | `internal/skill/skill.go` | 99 |
| `create skill directory: %w` | `Install` | `internal/skill/skill.go` | 103 |
| `write skill file: %w` | `installDedicated` | `internal/skill/skill.go` | 118 |
| `read %s: %w` | `installShared` | `internal/skill/skill.go` | 129 |
| `write skill file: %w` | `installShared` | `internal/skill/skill.go` | 134 |
| `update skill file: %w` | `installShared` | `internal/skill/skill.go` | 148 |
| `append skill file: %w` | `installShared` | `internal/skill/skill.go` | 162 |
| `open sqlite: %w` | `NewSQLiteStore` | `internal/store/sqlite.go` | 90 |
| `apply schema: %w` | `NewSQLiteStore` | `internal/store/sqlite.go` | 97 |
| `run migrations: %w` | `NewSQLiteStore` | `internal/store/sqlite.go` | 102 |
| `migration %q: %w` | `runMigrations` | `internal/store/sqlite.go` | 116 |
| `insert usage entry: %w` | `(*SQLiteStore).LogUsage` | `internal/store/sqlite.go` | 143 |
| `marshal labels: %w` | `(*SQLiteStore).UpsertIssueCache` | `internal/store/sqlite.go` | 157 |
| `upsert issue cache: %w` | `(*SQLiteStore).UpsertIssueCache` | `internal/store/sqlite.go` | 172 |
| `get issue cache: %w` | `(*SQLiteStore).GetIssueCache` | `internal/store/sqlite.go` | 193 |
| `list entries: %w` | `(*SQLiteStore).ListEntries` | `internal/store/sqlite.go` | 249 |
| `scan entry: %w` | `(*SQLiteStore).ListEntries` | `internal/store/sqlite.go` | 262 |
| `parse entry at: %w` | `(*SQLiteStore).ListEntries` | `internal/store/sqlite.go` | 267 |
| `iterate entries: %w` | `(*SQLiteStore).ListEntries` | `internal/store/sqlite.go` | 282 |
| `insert session: %w` | `(*SQLiteStore).StartSession` | `internal/store/sqlite.go` | 301 |
| `end session: %w` | `(*SQLiteStore).EndSession` | `internal/store/sqlite.go` | 315 |
| `end session rows affected: %w` | `(*SQLiteStore).EndSession` | `internal/store/sqlite.go` | 320 |
| `update session activity: %w` | `(*SQLiteStore).UpdateSessionActivity` | `internal/store/sqlite.go` | 336 |
| `end session: %w` | `(*SQLiteStore).EndSessionAt` | `internal/store/sqlite.go` | 348 |
| `end session rows affected: %w` | `(*SQLiteStore).EndSessionAt` | `internal/store/sqlite.go` | 353 |
| `list sessions: %w` | `(*SQLiteStore).ListSessions` | `internal/store/sqlite.go` | 372 |
| `scan session: %w` | `(*SQLiteStore).ListSessions` | `internal/store/sqlite.go` | 388 |
| `parse session started_at: %w` | `(*SQLiteStore).ListSessions` | `internal/store/sqlite.go` | 393 |
| `parse session ended_at: %w` | `(*SQLiteStore).ListSessions` | `internal/store/sqlite.go` | 399 |
| `iterate sessions: %w` | `(*SQLiteStore).ListSessions` | `internal/store/sqlite.go` | 434 |
| `fetch session tags: %w` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 455 |
| `marshal tags: %w` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 483 |
| `update session annotations: %w` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 499 |
| `list issues: %w` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 544 |
| `scan issue: %w` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 561 |
| `iterate issues: %w` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 588 |
| `get report: %w` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 611 |
| `scan report row: %w` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 623 |
| `iterate report rows: %w` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 636 |
| `list usage over time: %w` | `(*SQLiteStore).ListUsageOverTime` | `internal/store/sqlite.go` | 693 |
| `scan usage point: %w` | `(*SQLiteStore).ListUsageOverTime` | `internal/store/sqlite.go` | 714 |
| `iterate usage points: %w` | `(*SQLiteStore).ListUsageOverTime` | `internal/store/sqlite.go` | 734 |
| `list tracked issue refs: %w` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 756 |
| `scan tracked issue ref: %w` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 765 |
| `iterate tracked issue refs: %w` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 772 |
| `set budget: %w` | `(*SQLiteStore).SetBudget` | `internal/store/sqlite.go` | 830 |
| `unset budget: %w` | `(*SQLiteStore).UnsetBudget` | `internal/store/sqlite.go` | 842 |
| `get budget: %w` | `(*SQLiteStore).GetBudget` | `internal/store/sqlite.go` | 860 |
| `open browser: %w` | `openBrowserCmd` | `internal/tui/tui.go` | 739 |
| `` | `newTUITestStore` | `internal/tui/tui_test.go` | 43 |
| `` | `newTUIModel` | `internal/tui/tui_test.go` | 55 |
| `` | `newTUIModel` | `internal/tui/tui_test.go` | 57 |

