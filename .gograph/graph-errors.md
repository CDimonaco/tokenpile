# Errors & Panics

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-07-05 15:22:39 UTC  

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
| `read file: %w` | `runVerify` | `cmd/tokenpile/cmd_export.go` | 86 |
| `parse export: %w` | `runVerify` | `cmd/tokenpile/cmd_export.go` | 91 |
| `parse --pubkey: %w` | `runVerify` | `cmd/tokenpile/cmd_export.go` | 99 |
| `decode document public key: %w` | `runVerify` | `cmd/tokenpile/cmd_export.go` | 104 |
| `verification failed: document is not signed by the expected key` | `runVerify` | `cmd/tokenpile/cmd_export.go` | 110 |
| `verification failed: %w` | `runVerify` | `cmd/tokenpile/cmd_export.go` | 120 |
| `read pubkey file: %w` | `parseExpectedPubKey` | `cmd/tokenpile/cmd_export.go` | 153 |
| `invalid public key size: got %d, want %d` | `parseExpectedPubKey` | `cmd/tokenpile/cmd_export.go` | 158 |
| `decode pubkey file: %w` | `parseExpectedPubKey` | `cmd/tokenpile/cmd_export.go` | 166 |
| `invalid public key size: got %d, want %d` | `parseExpectedPubKey` | `cmd/tokenpile/cmd_export.go` | 170 |
| `list sessions: %w` | `gatherSessionsAndBudgets` | `cmd/tokenpile/cmd_export.go` | 187 |
| `list budgets: %w` | `gatherSessionsAndBudgets` | `cmd/tokenpile/cmd_export.go` | 192 |
| `parse --from: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 238 |
| `parse --to: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 247 |
| `list entries: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 255 |
| `build export: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 265 |
| `marshal export: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 270 |
| `write output: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 275 |
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
| `` | `newTestStore` | `cmd/tokenpile/integration_test.go` | 42 |
| `tui: %w` | `main` | `cmd/tokenpile/main.go` | 112 |
| `generate ed25519 key: %w` | `generateIdentity` | `internal/config/identity.go` | 24 |
| `write identity key: %w` | `generateIdentity` | `internal/config/identity.go` | 38 |
| `write identity pub: %w` | `generateIdentity` | `internal/config/identity.go` | 42 |
| `read identity key: %w` | `loadIdentity` | `internal/config/identity.go` | 53 |
| `decode identity key PEM` | `loadIdentity` | `internal/config/identity.go` | 58 |
| `invalid identity key size: got %d, want %d` | `loadIdentity` | `internal/config/identity.go` | 62 |
| `private key is not ed25519` | `loadIdentity` | `internal/config/identity.go` | 69 |
| `private key is not ed25519` | `Build` | `internal/export/export.go` | 110 |
| `canonical JSON: %w` | `documentDigest` | `internal/export/export.go` | 164 |
| `decode public key: %w` | `Verify` | `internal/export/export.go` | 177 |
| `invalid public key size: got %d, want %d` | `Verify` | `internal/export/export.go` | 181 |
| `decode signature: %w` | `Verify` | `internal/export/export.go` | 188 |
| `canonical JSON: %w` | `Verify` | `internal/export/export.go` | 204 |
| `unsupported schema version %q` | `Verify` | `internal/export/export.go` | 210 |
| `signature invalid: entries have been tampered with` | `Verify` | `internal/export/export.go` | 215 |
| `signature invalid: document has been tampered with` | `Verify` | `internal/export/export.go` | 218 |
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
| `start callback server: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 60 |
| `unexpected listener address type %T` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 66 |
| `oauth callback missing code` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 97 |
| `oauth callback: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 131 |
| `login timed out, please try again` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 134 |
| `exchange oauth code: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 144 |
| `store token: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 148 |
| `delete token from keychain: %w` | `(*GitHubAuthProvider).Logout` | `internal/provider/github_auth.go` | 174 |
| `remove credentials file: %w` | `(*GitHubAuthProvider).Logout` | `internal/provider/github_auth.go` | 181 |
| `create cipher: %w` | `(*GitHubAuthProvider).storeEncryptedToken` | `internal/provider/github_auth.go` | 203 |
| `create GCM: %w` | `(*GitHubAuthProvider).storeEncryptedToken` | `internal/provider/github_auth.go` | 208 |
| `generate nonce: %w` | `(*GitHubAuthProvider).storeEncryptedToken` | `internal/provider/github_auth.go` | 213 |
| `write credentials: %w` | `(*GitHubAuthProvider).storeEncryptedToken` | `internal/provider/github_auth.go` | 219 |
| `read credentials: %w` | `(*GitHubAuthProvider).loadEncryptedToken` | `internal/provider/github_auth.go` | 228 |
| `create cipher: %w` | `(*GitHubAuthProvider).loadEncryptedToken` | `internal/provider/github_auth.go` | 235 |
| `create GCM: %w` | `(*GitHubAuthProvider).loadEncryptedToken` | `internal/provider/github_auth.go` | 240 |
| `credentials file corrupted` | `(*GitHubAuthProvider).loadEncryptedToken` | `internal/provider/github_auth.go` | 245 |
| `decrypt credentials: %w` | `(*GitHubAuthProvider).loadEncryptedToken` | `internal/provider/github_auth.go` | 252 |
| `unsupported platform: %s` | `openBrowser` | `internal/provider/github_auth.go` | 302 |
| `id` | `TestEncryptedTokenRoundtrip` | `internal/provider/github_auth_internal_test.go` | 14 |
| `id` | `TestLoadEncryptedToken_CorruptedFile` | `internal/provider/github_auth_internal_test.go` | 25 |
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
| `%w: %s` | `Uninstall` | `internal/skill/skill.go` | 176 |
| `cannot determine install path for agent %s` | `Uninstall` | `internal/skill/skill.go` | 181 |
| `remove skill file: %w` | `uninstallDedicated` | `internal/skill/skill.go` | 197 |
| `read %s: %w` | `uninstallShared` | `internal/skill/skill.go` | 210 |
| `remove skill file: %w` | `uninstallShared` | `internal/skill/skill.go` | 225 |
| `update skill file: %w` | `uninstallShared` | `internal/skill/skill.go` | 234 |
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
| `list all sessions: %w` | `(*SQLiteStore).ListAllSessions` | `internal/store/sqlite.go` | 386 |
| `scan session: %w` | `scanSessions` | `internal/store/sqlite.go` | 407 |
| `parse session started_at: %w` | `scanSessions` | `internal/store/sqlite.go` | 412 |
| `parse session ended_at: %w` | `scanSessions` | `internal/store/sqlite.go` | 418 |
| `iterate sessions: %w` | `scanSessions` | `internal/store/sqlite.go` | 453 |
| `fetch session tags: %w` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 474 |
| `marshal tags: %w` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 502 |
| `update session annotations: %w` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 518 |
| `list issues: %w` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 563 |
| `scan issue: %w` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 576 |
| `iterate issues: %w` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 603 |
| `get report: %w` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 628 |
| `scan report row: %w` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 640 |
| `iterate report rows: %w` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 653 |
| `list usage over time: %w` | `(*SQLiteStore).ListUsageOverTime` | `internal/store/sqlite.go` | 710 |
| `scan usage point: %w` | `(*SQLiteStore).ListUsageOverTime` | `internal/store/sqlite.go` | 731 |
| `iterate usage points: %w` | `(*SQLiteStore).ListUsageOverTime` | `internal/store/sqlite.go` | 751 |
| `list tracked issue refs: %w` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 773 |
| `scan tracked issue ref: %w` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 782 |
| `iterate tracked issue refs: %w` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 789 |
| `set budget: %w` | `(*SQLiteStore).SetBudget` | `internal/store/sqlite.go` | 907 |
| `unset budget: %w` | `(*SQLiteStore).UnsetBudget` | `internal/store/sqlite.go` | 919 |
| `list budgets: %w` | `(*SQLiteStore).ListBudgets` | `internal/store/sqlite.go` | 930 |
| `scan budget: %w` | `(*SQLiteStore).ListBudgets` | `internal/store/sqlite.go` | 939 |
| `iterate budgets: %w` | `(*SQLiteStore).ListBudgets` | `internal/store/sqlite.go` | 946 |
| `get budget: %w` | `(*SQLiteStore).GetBudget` | `internal/store/sqlite.go` | 964 |
| `open browser: %w` | `openBrowserCmd` | `internal/tui/tui.go` | 739 |
| `` | `newTUITestStore` | `internal/tui/tui_test.go` | 43 |
| `` | `newTUIModel` | `internal/tui/tui_test.go` | 55 |
| `` | `newTUIModel` | `internal/tui/tui_test.go` | 57 |

