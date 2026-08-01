# Errors & Panics

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-08-01 13:49:30 UTC  

---

| Message | Function | File | Line |
|---------|----------|------|------|
| `logout: %w` | `authCommands` | `cmd/tokenpile/cmd_auth.go` | 79 |
| `--use-gh-cli and --no-gh-cli are mutually exclusive` | `runLogin` | `cmd/tokenpile/cmd_auth.go` | 106 |
| `login: %w` | `runLogin` | `cmd/tokenpile/cmd_auth.go` | 114 |
| `read gh credential: %w` | `loginWithGhCli` | `cmd/tokenpile/cmd_auth.go` | 164 |
| `validate gh credential: %w` | `loginWithGhCli` | `cmd/tokenpile/cmd_auth.go` | 171 |
| `gh credential for %s lacks the repo scope needed to read issues: run gh auth refresh -s repo` | `loginWithGhCli` | `cmd/tokenpile/cmd_auth.go` | 175 |
| `gh CLI token source unavailable: %w` | `ghStatus` | `cmd/tokenpile/cmd_auth.go` | 219 |
| `oauth failed` | `TestAuthLogin_OAuthFailure` | `cmd/tokenpile/cmd_auth_test.go` | 101 |
| `401 bad credentials` | `TestAuthLogin_ValidationFailureDoesNotPersist` | `cmd/tokenpile/cmd_auth_test.go` | 217 |
| `not found` | `TestAuthStatus_NotLoggedIn` | `cmd/tokenpile/cmd_auth_test.go` | 239 |
| `resolve repo: %w` | `bindCommand` | `cmd/tokenpile/cmd_bind.go` | 55 |
| `resolve working directory: %w` | `bindCommand` | `cmd/tokenpile/cmd_bind.go` | 60 |
| `bind issue: %w` | `bindCommand` | `cmd/tokenpile/cmd_bind.go` | 71 |
| `session id is required` | `unattributedCommand` | `cmd/tokenpile/cmd_bind.go` | 111 |
| `assign issue: %w` | `unattributedCommand` | `cmd/tokenpile/cmd_bind.go` | 121 |
| `session id is required` | `unattributedCommand` | `cmd/tokenpile/cmd_bind.go` | 136 |
| `unassign issue: %w` | `unattributedCommand` | `cmd/tokenpile/cmd_bind.go` | 141 |
| `list unattributed: %w` | `unattributedCommand` | `cmd/tokenpile/cmd_bind.go` | 155 |
| `--issue must be a positive issue number` | `resolveAssignIssue` | `cmd/tokenpile/cmd_bind.go` | 190 |
| `list unattributed: %w` | `resolveAssignIssue` | `cmd/tokenpile/cmd_bind.go` | 198 |
| `no issue suggested for this session: pass --issue` | `resolveAssignIssue` | `cmd/tokenpile/cmd_bind.go` | 207 |
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
| `read pubkey file: %w` | `parseExpectedPubKey` | `cmd/tokenpile/cmd_export.go` | 146 |
| `invalid public key size: got %d, want %d` | `parseExpectedPubKey` | `cmd/tokenpile/cmd_export.go` | 151 |
| `decode pubkey file: %w` | `parseExpectedPubKey` | `cmd/tokenpile/cmd_export.go` | 159 |
| `invalid public key size: got %d, want %d` | `parseExpectedPubKey` | `cmd/tokenpile/cmd_export.go` | 163 |
| `list sessions: %w` | `gatherSessionsAndBudgets` | `cmd/tokenpile/cmd_export.go` | 180 |
| `list budgets: %w` | `gatherSessionsAndBudgets` | `cmd/tokenpile/cmd_export.go` | 185 |
| `parse --from: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 242 |
| `parse --to: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 251 |
| `list entries: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 259 |
| `build export: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 269 |
| `marshal export: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 274 |
| `write output: %w` | `runExport` | `cmd/tokenpile/cmd_export.go` | 279 |
| `agent name is required: tokenpile hook <agent>` | `hookCommand` | `cmd/tokenpile/cmd_hook.go` | 37 |
| `read hook payload: %w` | `hookCommand` | `cmd/tokenpile/cmd_hook.go` | 42 |
| `read %s payload: %w` | `hookCommand` | `cmd/tokenpile/cmd_hook.go` | 59 |
| `spool turns: %w` | `hookCommand` | `cmd/tokenpile/cmd_hook.go` | 63 |
| `parse hook payload: %w` | `turnsFromPayload` | `cmd/tokenpile/cmd_hook.go` | 76 |
| `hook payload carries no transcript_path` | `turnsFromPayload` | `cmd/tokenpile/cmd_hook.go` | 80 |
| `unsupported agent %q` | `turnsFromPayload` | `cmd/tokenpile/cmd_hook.go` | 87 |
| `cannot infer repo: pass --repo owner/repo or run from inside a GitHub repository` | `runLog` | `cmd/tokenpile/cmd_log.go` | 86 |
| `infer repo: %w` | `runLog` | `cmd/tokenpile/cmd_log.go` | 91 |
| `%s must be zero or greater` | `runLog` | `cmd/tokenpile/cmd_log.go` | 112 |
| `at least one token count is required: --input, --cache-write, --cache-read or --output` | `runLog` | `cmd/tokenpile/cmd_log.go` | 117 |
| `--reasoning cannot exceed --output: reasoning tokens are a subset of output` | `runLog` | `cmd/tokenpile/cmd_log.go` | 123 |
| `resolve session: %w` | `runLog` | `cmd/tokenpile/cmd_log.go` | 141 |
| `log usage: %w` | `runLog` | `cmd/tokenpile/cmd_log.go` | 159 |
| `--issue must be a positive issue number` | `validateAndCacheIssue` | `cmd/tokenpile/cmd_log.go` | 186 |
| `issue #%d not found in %s` | `validateAndCacheIssue` | `cmd/tokenpile/cmd_log.go` | 192 |
| `GitHub authentication required to validate issues: run tokenpile auth login` | `validateAndCacheIssue` | `cmd/tokenpile/cmd_log.go` | 196 |
| `validate issue: %w` | `validateAndCacheIssue` | `cmd/tokenpile/cmd_log.go` | 199 |
| `list sessions: %w` | `resolveSession` | `cmd/tokenpile/cmd_log.go` | 236 |
| `end idle session: %w` | `resolveSession` | `cmd/tokenpile/cmd_log.go` | 250 |
| `start session: %w` | `resolveSession` | `cmd/tokenpile/cmd_log.go` | 267 |
| `model name is required` | `pricingCommands` | `cmd/tokenpile/cmd_pricing.go` | 74 |
| `set pricing: %w` | `pricingCommands` | `cmd/tokenpile/cmd_pricing.go` | 85 |
| `cannot infer repo: pass --repo owner/repo or run from inside a GitHub repository` | `reportCommand` | `cmd/tokenpile/cmd_report.go` | 52 |
| `infer repo: %w` | `reportCommand` | `cmd/tokenpile/cmd_report.go` | 57 |
| `get report: %w` | `reportCommand` | `cmd/tokenpile/cmd_report.go` | 69 |
| `get issue cache: %w` | `reportCommand` | `cmd/tokenpile/cmd_report.go` | 74 |
| `get budget: %w` | `reportCommand` | `cmd/tokenpile/cmd_report.go` | 96 |
| `list sessions: %w` | `printSessionsReport` | `cmd/tokenpile/cmd_report.go` | 115 |
| `reset aborted` | `runReset` | `cmd/tokenpile/cmd_reset.go` | 77 |
| `backup failed, nothing was deleted: %w` | `runReset` | `cmd/tokenpile/cmd_reset.go` | 83 |
| `reset incomplete: %d item(s) could not be removed` | `runReset` | `cmd/tokenpile/cmd_reset.go` | 95 |
| `list entries: %w` | `writeResetBackup` | `cmd/tokenpile/cmd_reset.go` | 161 |
| `list sessions: %w` | `writeResetBackup` | `cmd/tokenpile/cmd_reset.go` | 166 |
| `list budgets: %w` | `writeResetBackup` | `cmd/tokenpile/cmd_reset.go` | 171 |
| `build export: %w` | `writeResetBackup` | `cmd/tokenpile/cmd_reset.go` | 181 |
| `marshal export: %w` | `writeResetBackup` | `cmd/tokenpile/cmd_reset.go` | 186 |
| `write backup: %w` | `writeResetBackup` | `cmd/tokenpile/cmd_reset.go` | 195 |
| `logout: %w` | `destroyState` | `cmd/tokenpile/cmd_reset.go` | 210 |
| `uninstall skill %s: %w` | `destroyState` | `cmd/tokenpile/cmd_reset.go` | 218 |
| `remove %s: %w` | `destroyState` | `cmd/tokenpile/cmd_reset.go` | 233 |
| `install skill: %w` | `skillCommands` | `cmd/tokenpile/cmd_skill.go` | 32 |
| `tui: %w` | `main` | `cmd/tokenpile/main.go` | 128 |
| `read bindings: %w` | `(*Store).load` | `internal/attribution/attribution.go` | 54 |
| `parse bindings: %w` | `(*Store).load` | `internal/attribution/attribution.go` | 58 |
| `create binding directory: %w` | `(*Store).save` | `internal/attribution/attribution.go` | 74 |
| `marshal bindings: %w` | `(*Store).save` | `internal/attribution/attribution.go` | 79 |
| `write bindings: %w` | `(*Store).save` | `internal/attribution/attribution.go` | 83 |
| `read transcript: %w` | `ReadClaudeCodeTranscript` | `internal/capture/claudecode.go` | 116 |
| `transcript unparseable: %d malformed lines` | `ReadClaudeCodeTranscript` | `internal/capture/claudecode.go` | 120 |
| `open transcript: %w` | `ReadClaudeCodeTranscriptFile` | `internal/capture/claudecode.go` | 134 |
| `decode opencode payload: %w` | `ReadOpenCodePayload` | `internal/capture/opencode.go` | 60 |
| `create spool directory: %w` | `(*Spool).Append` | `internal/capture/spool.go` | 40 |
| `open spool: %w` | `(*Spool).Append` | `internal/capture/spool.go` | 45 |
| `write spool record: %w` | `(*Spool).Append` | `internal/capture/spool.go` | 52 |
| `create spool directory: %w` | `(*Spool).AppendRaw` | `internal/capture/spool.go` | 63 |
| `open raw spool: %w` | `(*Spool).AppendRaw` | `internal/capture/spool.go` | 70 |
| `open spool: %w` | `(*Spool).Read` | `internal/capture/spool.go` | 94 |
| `read spool: %w` | `(*Spool).Read` | `internal/capture/spool.go` | 113 |
| `clear spool: %w` | `(*Spool).Clear` | `internal/capture/spool.go` | 124 |
| `generate ed25519 key: %w` | `generateIdentity` | `internal/config/identity.go` | 24 |
| `write identity key: %w` | `generateIdentity` | `internal/config/identity.go` | 38 |
| `write identity pub: %w` | `generateIdentity` | `internal/config/identity.go` | 42 |
| `read identity key: %w` | `loadIdentity` | `internal/config/identity.go` | 53 |
| `decode identity key PEM` | `loadIdentity` | `internal/config/identity.go` | 58 |
| `invalid identity key size: got %d, want %d` | `loadIdentity` | `internal/config/identity.go` | 62 |
| `private key is not ed25519` | `loadIdentity` | `internal/config/identity.go` | 69 |
| `private key is not ed25519` | `Build` | `internal/export/export.go` | 112 |
| `canonical JSON: %w` | `documentDigest` | `internal/export/export.go` | 166 |
| `decode public key: %w` | `Verify` | `internal/export/export.go` | 179 |
| `invalid public key size: got %d, want %d` | `Verify` | `internal/export/export.go` | 183 |
| `decode signature: %w` | `Verify` | `internal/export/export.go` | 190 |
| `unsupported schema version %q: only %s can be verified` | `Verify` | `internal/export/export.go` | 208 |
| `signature invalid: document has been tampered with` | `Verify` | `internal/export/export.go` | 215 |
| `parse default pricing: %w` | `NewLoader` | `internal/pricing/pricing.go` | 40 |
| `read pricing override: %w` | `NewLoader` | `internal/pricing/pricing.go` | 49 |
| `parse pricing override: %w` | `NewLoader` | `internal/pricing/pricing.go` | 55 |
| `read pricing override: %w` | `(*Loader).SetOverride` | `internal/pricing/pricing.go` | 164 |
| `parse pricing override: %w` | `(*Loader).SetOverride` | `internal/pricing/pricing.go` | 170 |
| `marshal pricing override: %w` | `(*Loader).SetOverride` | `internal/pricing/pricing.go` | 182 |
| `write pricing override: %w` | `(*Loader).SetOverride` | `internal/pricing/pricing.go` | 186 |
| `run gh %s: %w` | `(*GhCliAuthProvider).run` | `internal/provider/ghcli_auth.go` | 119 |
| `%w (gh said: %s)` | `withDetail` | `internal/provider/ghcli_auth.go` | 130 |
| `start callback server: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 60 |
| `unexpected listener address type %T` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 66 |
| `oauth callback missing code` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 97 |
| `oauth callback: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 131 |
| `login timed out, please try again` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 134 |
| `exchange oauth code: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 144 |
| `store token: %w` | `(*GitHubAuthProvider).Login` | `internal/provider/github_auth.go` | 148 |
| `delete token from keychain: %w` | `(*GitHubAuthProvider).Logout` | `internal/provider/github_auth.go` | 176 |
| `remove credentials file: %w` | `(*GitHubAuthProvider).Logout` | `internal/provider/github_auth.go` | 183 |
| `create cipher: %w` | `storeEncryptedToken` | `internal/provider/github_auth.go` | 224 |
| `create GCM: %w` | `storeEncryptedToken` | `internal/provider/github_auth.go` | 229 |
| `generate nonce: %w` | `storeEncryptedToken` | `internal/provider/github_auth.go` | 234 |
| `write credentials: %w` | `storeEncryptedToken` | `internal/provider/github_auth.go` | 240 |
| `read credentials: %w` | `loadEncryptedToken` | `internal/provider/github_auth.go` | 249 |
| `create cipher: %w` | `loadEncryptedToken` | `internal/provider/github_auth.go` | 256 |
| `create GCM: %w` | `loadEncryptedToken` | `internal/provider/github_auth.go` | 261 |
| `credentials file corrupted` | `loadEncryptedToken` | `internal/provider/github_auth.go` | 266 |
| `decrypt credentials: %w` | `loadEncryptedToken` | `internal/provider/github_auth.go` | 273 |
| `unsupported platform: %s` | `openBrowser` | `internal/provider/github_auth.go` | 323 |
| `set base URL: %w` | `(*GitHubIssueProvider).client` | `internal/provider/github_issues.go` | 45 |
| `%s: %w` | `(*GitHubIssueProvider).wrapAccessError` | `internal/provider/github_issues.go` | 60 |
| `%s: %w (the gh CLI credential does not grant access to this repository)` | `(*GitHubIssueProvider).wrapAccessError` | `internal/provider/github_issues.go` | 64 |
| `invalid repo format %q: expected owner/repo` | `(*GitHubIssueProvider).ListIssues` | `internal/provider/github_issues.go` | 83 |
| `invalid repo format %q: expected owner/repo` | `(*GitHubIssueProvider).GetIssue` | `internal/provider/github_issues.go` | 144 |
| `cannot infer repo from remote %q: not a GitHub remote; pass --repo owner/repo` | `ParseRemote` | `internal/provider/repoinfer.go` | 58 |
| `persist token source: %w` | `PersistGhCliTokenSource` | `internal/provider/tokensource.go` | 49 |
| `set base URL: %w` | `ValidateTokenWithURL` | `internal/provider/validate.go` | 52 |
| `validate token: %w` | `ValidateTokenWithURL` | `internal/provider/validate.go` | 58 |
| `remove plugin: %w` | `UninstallHook` | `internal/skill/hooks.go` | 81 |
| `read settings: %w` | `readSettings` | `internal/skill/hooks.go` | 109 |
| `parse settings: %w` | `readSettings` | `internal/skill/hooks.go` | 118 |
| `create settings directory: %w` | `writeSettings` | `internal/skill/hooks.go` | 126 |
| `marshal settings: %w` | `writeSettings` | `internal/skill/hooks.go` | 131 |
| `write settings: %w` | `writeSettings` | `internal/skill/hooks.go` | 135 |
| `marshal hook entry: %w` | `toGeneric` | `internal/skill/hooks.go` | 245 |
| `normalize hook entry: %w` | `toGeneric` | `internal/skill/hooks.go` | 250 |
| `create plugin directory: %w` | `installOpenCodePlugin` | `internal/skill/hooks.go` | 261 |
| `write plugin: %w` | `installOpenCodePlugin` | `internal/skill/hooks.go` | 265 |
| `%w: %s` | `Install` | `internal/skill/skill.go` | 102 |
| `cannot determine install path for agent %s` | `Install` | `internal/skill/skill.go` | 107 |
| `create skill directory: %w` | `Install` | `internal/skill/skill.go` | 111 |
| `install capture hook: %w` | `Install` | `internal/skill/skill.go` | 119 |
| `write skill file: %w` | `installDedicated` | `internal/skill/skill.go` | 146 |
| `%w: %s` | `Uninstall` | `internal/skill/skill.go` | 160 |
| `cannot determine install path for agent %s` | `Uninstall` | `internal/skill/skill.go` | 165 |
| `remove capture hook: %w` | `Uninstall` | `internal/skill/skill.go` | 171 |
| `remove skill file: %w` | `uninstallDedicated` | `internal/skill/skill.go` | 183 |
| `read %s: %w` | `uninstallShared` | `internal/skill/skill.go` | 199 |
| `remove skill file: %w` | `uninstallShared` | `internal/skill/skill.go` | 214 |
| `update skill file: %w` | `uninstallShared` | `internal/skill/skill.go` | 223 |
| `open sqlite: %w` | `NewSQLiteStore` | `internal/store/sqlite.go` | 102 |
| `apply schema: %w` | `NewSQLiteStore` | `internal/store/sqlite.go` | 109 |
| `run migrations: %w` | `NewSQLiteStore` | `internal/store/sqlite.go` | 114 |
| `migration %q: %w` | `runMigrations` | `internal/store/sqlite.go` | 128 |
| `insert usage entry: %w` | `(*SQLiteStore).LogUsage` | `internal/store/sqlite.go` | 161 |
| `marshal labels: %w` | `(*SQLiteStore).UpsertIssueCache` | `internal/store/sqlite.go` | 175 |
| `upsert issue cache: %w` | `(*SQLiteStore).UpsertIssueCache` | `internal/store/sqlite.go` | 190 |
| `get issue cache: %w` | `(*SQLiteStore).GetIssueCache` | `internal/store/sqlite.go` | 211 |
| `list entries: %w` | `(*SQLiteStore).ListEntries` | `internal/store/sqlite.go` | 268 |
| `scan entry: %w` | `(*SQLiteStore).ListEntries` | `internal/store/sqlite.go` | 286 |
| `parse entry at: %w` | `(*SQLiteStore).ListEntries` | `internal/store/sqlite.go` | 298 |
| `iterate entries: %w` | `(*SQLiteStore).ListEntries` | `internal/store/sqlite.go` | 313 |
| `insert session: %w` | `(*SQLiteStore).StartSession` | `internal/store/sqlite.go` | 332 |
| `end session: %w` | `(*SQLiteStore).EndSession` | `internal/store/sqlite.go` | 346 |
| `end session rows affected: %w` | `(*SQLiteStore).EndSession` | `internal/store/sqlite.go` | 351 |
| `update session activity: %w` | `(*SQLiteStore).UpdateSessionActivity` | `internal/store/sqlite.go` | 367 |
| `end session: %w` | `(*SQLiteStore).EndSessionAt` | `internal/store/sqlite.go` | 379 |
| `end session rows affected: %w` | `(*SQLiteStore).EndSessionAt` | `internal/store/sqlite.go` | 384 |
| `list sessions: %w` | `(*SQLiteStore).ListSessions` | `internal/store/sqlite.go` | 410 |
| `list all sessions: %w` | `(*SQLiteStore).ListAllSessions` | `internal/store/sqlite.go` | 424 |
| `scan session: %w` | `scanSessions` | `internal/store/sqlite.go` | 445 |
| `parse session started_at: %w` | `scanSessions` | `internal/store/sqlite.go` | 450 |
| `parse session ended_at: %w` | `scanSessions` | `internal/store/sqlite.go` | 456 |
| `iterate sessions: %w` | `scanSessions` | `internal/store/sqlite.go` | 491 |
| `fetch session tags: %w` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 512 |
| `marshal tags: %w` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 540 |
| `update session annotations: %w` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 556 |
| `list issues: %w` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 601 |
| `scan issue: %w` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 617 |
| `iterate issues: %w` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 643 |
| `get report: %w` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 668 |
| `scan report row: %w` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 682 |
| `iterate report rows: %w` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 698 |
| `list usage over time: %w` | `(*SQLiteStore).ListUsageOverTime` | `internal/store/sqlite.go` | 755 |
| `scan usage point: %w` | `(*SQLiteStore).ListUsageOverTime` | `internal/store/sqlite.go` | 776 |
| `iterate usage points: %w` | `(*SQLiteStore).ListUsageOverTime` | `internal/store/sqlite.go` | 795 |
| `list tracked issue refs: %w` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 816 |
| `scan tracked issue ref: %w` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 825 |
| `iterate tracked issue refs: %w` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 832 |
| `set budget: %w` | `(*SQLiteStore).SetBudget` | `internal/store/sqlite.go` | 950 |
| `unset budget: %w` | `(*SQLiteStore).UnsetBudget` | `internal/store/sqlite.go` | 962 |
| `list budgets: %w` | `(*SQLiteStore).ListBudgets` | `internal/store/sqlite.go` | 973 |
| `scan budget: %w` | `(*SQLiteStore).ListBudgets` | `internal/store/sqlite.go` | 982 |
| `iterate budgets: %w` | `(*SQLiteStore).ListBudgets` | `internal/store/sqlite.go` | 989 |
| `get budget: %w` | `(*SQLiteStore).GetBudget` | `internal/store/sqlite.go` | 1007 |
| `list unattributed: %w` | `(*SQLiteStore).ListUnattributed` | `internal/store/sqlite.go` | 1036 |
| `scan unattributed: %w` | `(*SQLiteStore).ListUnattributed` | `internal/store/sqlite.go` | 1060 |
| `iterate unattributed: %w` | `(*SQLiteStore).ListUnattributed` | `internal/store/sqlite.go` | 1098 |
| `session id is required` | `(*SQLiteStore).setSessionIssue` | `internal/store/sqlite.go` | 1131 |
| `begin transaction: %w` | `(*SQLiteStore).setSessionIssue` | `internal/store/sqlite.go` | 1136 |
| `update entries: %w` | `(*SQLiteStore).setSessionIssue` | `internal/store/sqlite.go` | 1143 |
| `rows affected: %w` | `(*SQLiteStore).setSessionIssue` | `internal/store/sqlite.go` | 1148 |
| `update session: %w` | `(*SQLiteStore).setSessionIssue` | `internal/store/sqlite.go` | 1153 |
| `commit: %w` | `(*SQLiteStore).setSessionIssue` | `internal/store/sqlite.go` | 1157 |
| `open browser: %w` | `openBrowserCmd` | `internal/tui/tui.go` | 835 |
| `enter a positive issue number` | `(Model).handleAssignKey` | `internal/tui/unattributed.go` | 60 |

