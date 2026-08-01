# SQL Queries

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-08-01 14:05:29 UTC  

---

| Query | Function | File | Line |
|-------|----------|------|------|
| `UPDATE usage_entries 		SET input_fresh = tokens_in, output = tokens_out 		WHERE input_fresh = 0 AND output = 0` | `backfillLegacyTokenColumns` | `internal/store/sqlite.go` | 183 |
| `SELECT "notnull" FROM pragma_table_info(?) WHERE name = ?` | `inspectColumn` | `internal/store/sqlite.go` | 284 |
| `INSERT OR IGNORE INTO usage_entries 		 (id, repo, issue_num, agent, model, input_fresh, cache_write, cache_read, output, reasoning, source, session_id, branch, at) 		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)` | `(*SQLiteStore).LogUsage` | `internal/store/sqlite.go` | 355 |
| ` 		INSERT INTO issue_cache (repo, issue_num, title, labels, created_at, updated_at) 		VALUES (?, ?, ?, ?, ?, ?) 		ON CONFLICT(repo, issue_num) DO UPDATE SET 			title = excluded.title, 			labels = excluded.labels, 			updated_at = excluded.updated_at` | `(*SQLiteStore).UpsertIssueCache` | `internal/store/sqlite.go` | 387 |
| `SELECT repo, issue_num, title, labels, created_at, updated_at 		 FROM issue_cache WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).GetIssueCache` | `internal/store/sqlite.go` | 407 |
| `SELECT ue.id, ue.repo, ue.issue_num, ue.agent, ue.model, 		ue.input_fresh, ue.cache_write, ue.cache_read, ue.output, ue.reasoning, ue.source, 		ue.session_id, ue.branch, ue.at, 		COALESCE(ic.title, ''), COALESCE(ic.labels, '[]') 		FROM usage_entries ue 		LEFT JOIN issue_cache ic ON ic.repo = ue.repo AND ic.issue_num = ue.issue_num 		WHERE 1=1` | `(*SQLiteStore).ListEntries` | `internal/store/sqlite.go` | 473 |
| `INSERT INTO sessions (id, repo, issue_num, started_at) VALUES (?, ?, ?, ?)` | `(*SQLiteStore).StartSession` | `internal/store/sqlite.go` | 534 |
| `UPDATE sessions SET ended_at = ? WHERE id = ? AND ended_at IS NULL` | `(*SQLiteStore).EndSession` | `internal/store/sqlite.go` | 548 |
| `UPDATE sessions SET last_activity_at = ? WHERE id = ? AND ended_at IS NULL` | `(*SQLiteStore).UpdateSessionActivity` | `internal/store/sqlite.go` | 569 |
| `UPDATE sessions SET ended_at = ? WHERE id = ? AND ended_at IS NULL` | `(*SQLiteStore).EndSessionAt` | `internal/store/sqlite.go` | 581 |
| `SELECT id, repo, issue_num, started_at, ended_at, last_activity_at, note, tags 		 FROM sessions WHERE repo = ? AND issue_num IS NULL ORDER BY started_at` | `(*SQLiteStore).ListSessions` | `internal/store/sqlite.go` | 615 |
| `SELECT id, repo, issue_num, started_at, ended_at, last_activity_at, note, tags 		 FROM sessions ORDER BY repo, issue_num, started_at` | `(*SQLiteStore).ListAllSessions` | `internal/store/sqlite.go` | 625 |
| `SELECT tags FROM sessions WHERE id = ?` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 713 |
| `UPDATE sessions SET note = ?, tags = ? WHERE id = ?` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 751 |
| `UPDATE sessions SET tags = ? WHERE id = ?` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 756 |
| ` 		SELECT ue.repo, ue.issue_num, ue.model, SUM(ue.input_fresh), SUM(ue.cache_write), SUM(ue.cache_read), SUM(ue.output), SUM(ue.reasoning), 		       COALESCE(ic.title, ''), COALESCE(ic.labels, '[]'), ib.budget 		FROM usage_entries ue 		LEFT JOIN issue_cache ic ON ic.repo = ue.repo AND ic.issue_num = ue.issue_num 		LEFT JOIN issue_budgets ib ON ib.repo = ue.repo AND ib.issue_num = ue.issue_num 		WHERE ue.issue_num IS NOT NULL` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 806 |
| `SELECT agent, model, COUNT(*), SUM(input_fresh), SUM(cache_write), SUM(cache_read), SUM(output), SUM(reasoning) 		 FROM usage_entries 		 WHERE repo = ? AND issue_num = ? 		 GROUP BY agent, model 		 ORDER BY agent, model` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 866 |
| `SELECT DISTINCT repo, issue_num FROM usage_entries WHERE issue_num IS NOT NULL ORDER BY repo, issue_num` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 1019 |
| `SELECT repo, issue_num, started_at, ended_at FROM sessions` | `(*SQLiteStore).totalTimes` | `internal/store/sqlite.go` | 1061 |
| `SELECT started_at, ended_at FROM sessions WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).totalTime` | `internal/store/sqlite.go` | 1106 |
| `INSERT INTO issue_budgets (repo, issue_num, budget) VALUES (?, ?, ?) 		 ON CONFLICT(repo, issue_num) DO UPDATE SET budget = excluded.budget` | `(*SQLiteStore).SetBudget` | `internal/store/sqlite.go` | 1151 |
| `DELETE FROM issue_budgets WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).UnsetBudget` | `internal/store/sqlite.go` | 1164 |
| `SELECT repo, issue_num, budget FROM issue_budgets ORDER BY repo, issue_num` | `(*SQLiteStore).ListBudgets` | `internal/store/sqlite.go` | 1176 |
| `SELECT budget FROM issue_budgets WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).GetBudget` | `internal/store/sqlite.go` | 1204 |
| ` 		SELECT COALESCE(ue.session_id, ''), ue.repo, ue.branch, ue.model, COUNT(*), 		       SUM(ue.input_fresh), SUM(ue.cache_write), SUM(ue.cache_read), 		       SUM(ue.output), SUM(ue.reasoning), 		       MIN(ue.at), MAX(ue.at) 		FROM usage_entries ue 		WHERE ue.issue_num IS NULL` | `(*SQLiteStore).ListUnattributed` | `internal/store/sqlite.go` | 1241 |
| `UPDATE usage_entries SET issue_num = ? WHERE session_id = ?` | `(*SQLiteStore).setSessionIssue` | `internal/store/sqlite.go` | 1347 |
| `UPDATE sessions SET issue_num = ? WHERE id = ?` | `(*SQLiteStore).setSessionIssue` | `internal/store/sqlite.go` | 1358 |
| ` CREATE TABLE usage_entries (     id TEXT PRIMARY KEY, repo TEXT NOT NULL, issue_num INTEGER NOT NULL,     agent TEXT NOT NULL, model TEXT NOT NULL,     input_fresh INTEGER NOT NULL, cache_write INTEGER NOT NULL,     cache_read INTEGER NOT NULL, output INTEGER NOT NULL, reasoning INTEGER NOT NULL,     source TEXT NOT NULL, session_id TEXT, at TEXT NOT NULL); CREATE TABLE sessions (     id TEXT PRIMARY KEY, repo TEXT NOT NULL, issue_num INTEGER NOT NULL,     started_at TEXT NOT NULL, ended_at TEXT, note TEXT, tags TEXT, last_activity_at TEXT); INSERT INTO usage_entries VALUES     ('e1','o/r',42,'claude-code','m',10,0,0,5,0,'estimated','s1','2026-07-01T00:00:00Z'); INSERT INTO sessions VALUES     ('s1','o/r',42,'2026-07-01T00:00:00Z',NULL,NULL,NULL,'2026-07-01T00:00:00Z');` | `TestSQLiteStore_UpgradeFromNotNullIssueNum` | `internal/store/sqlite_test.go` | 763 |
| ` CREATE TABLE usage_entries (     id TEXT PRIMARY KEY, repo TEXT NOT NULL, issue_num INTEGER NOT NULL,     agent TEXT NOT NULL, model TEXT NOT NULL,     tokens_in INTEGER NOT NULL, tokens_out INTEGER NOT NULL,     session_id TEXT, at TEXT NOT NULL); CREATE TABLE sessions (     id TEXT PRIMARY KEY, repo TEXT NOT NULL, issue_num INTEGER NOT NULL,     started_at TEXT NOT NULL, ended_at TEXT, note TEXT, tags TEXT, last_activity_at TEXT); INSERT INTO usage_entries VALUES     ('e1','o/r',42,'claude-code','m',45000,3000,'s1','2026-07-01T00:00:00Z');` | `TestSQLiteStore_UpgradeFromLegacyTokenColumns` | `internal/store/sqlite_test.go` | 837 |

