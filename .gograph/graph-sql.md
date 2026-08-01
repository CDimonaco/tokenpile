# SQL Queries

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-08-01 13:38:20 UTC  

---

| Query | Function | File | Line |
|-------|----------|------|------|
| `INSERT OR IGNORE INTO usage_entries 		 (id, repo, issue_num, agent, model, input_fresh, cache_write, cache_read, output, reasoning, source, session_id, at) 		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)` | `(*SQLiteStore).LogUsage` | `internal/store/sqlite.go` | 142 |
| ` 		INSERT INTO issue_cache (repo, issue_num, title, labels, created_at, updated_at) 		VALUES (?, ?, ?, ?, ?, ?) 		ON CONFLICT(repo, issue_num) DO UPDATE SET 			title = excluded.title, 			labels = excluded.labels, 			updated_at = excluded.updated_at` | `(*SQLiteStore).UpsertIssueCache` | `internal/store/sqlite.go` | 174 |
| `SELECT repo, issue_num, title, labels, created_at, updated_at 		 FROM issue_cache WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).GetIssueCache` | `internal/store/sqlite.go` | 194 |
| `SELECT ue.id, ue.repo, ue.issue_num, ue.agent, ue.model, 		ue.input_fresh, ue.cache_write, ue.cache_read, ue.output, ue.reasoning, ue.source, 		ue.session_id, ue.at, 		COALESCE(ic.title, ''), COALESCE(ic.labels, '[]') 		FROM usage_entries ue 		LEFT JOIN issue_cache ic ON ic.repo = ue.repo AND ic.issue_num = ue.issue_num 		WHERE 1=1` | `(*SQLiteStore).ListEntries` | `internal/store/sqlite.go` | 260 |
| `INSERT INTO sessions (id, repo, issue_num, started_at) VALUES (?, ?, ?, ?)` | `(*SQLiteStore).StartSession` | `internal/store/sqlite.go` | 321 |
| `UPDATE sessions SET ended_at = ? WHERE id = ? AND ended_at IS NULL` | `(*SQLiteStore).EndSession` | `internal/store/sqlite.go` | 335 |
| `UPDATE sessions SET last_activity_at = ? WHERE id = ? AND ended_at IS NULL` | `(*SQLiteStore).UpdateSessionActivity` | `internal/store/sqlite.go` | 356 |
| `UPDATE sessions SET ended_at = ? WHERE id = ? AND ended_at IS NULL` | `(*SQLiteStore).EndSessionAt` | `internal/store/sqlite.go` | 368 |
| `SELECT id, repo, issue_num, started_at, ended_at, last_activity_at, note, tags 		 FROM sessions WHERE repo = ? AND issue_num IS NULL ORDER BY started_at` | `(*SQLiteStore).ListSessions` | `internal/store/sqlite.go` | 402 |
| `SELECT id, repo, issue_num, started_at, ended_at, last_activity_at, note, tags 		 FROM sessions ORDER BY repo, issue_num, started_at` | `(*SQLiteStore).ListAllSessions` | `internal/store/sqlite.go` | 412 |
| `SELECT tags FROM sessions WHERE id = ?` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 500 |
| `UPDATE sessions SET note = ?, tags = ? WHERE id = ?` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 538 |
| `UPDATE sessions SET tags = ? WHERE id = ?` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 543 |
| ` 		SELECT ue.repo, ue.issue_num, ue.model, SUM(ue.input_fresh), SUM(ue.cache_write), SUM(ue.cache_read), SUM(ue.output), SUM(ue.reasoning), 		       COALESCE(ic.title, ''), COALESCE(ic.labels, '[]'), ib.budget 		FROM usage_entries ue 		LEFT JOIN issue_cache ic ON ic.repo = ue.repo AND ic.issue_num = ue.issue_num 		LEFT JOIN issue_budgets ib ON ib.repo = ue.repo AND ib.issue_num = ue.issue_num 		WHERE ue.issue_num IS NOT NULL` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 593 |
| `SELECT agent, model, COUNT(*), SUM(input_fresh), SUM(cache_write), SUM(cache_read), SUM(output), SUM(reasoning) 		 FROM usage_entries 		 WHERE repo = ? AND issue_num = ? 		 GROUP BY agent, model 		 ORDER BY agent, model` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 653 |
| `SELECT DISTINCT repo, issue_num FROM usage_entries WHERE issue_num IS NOT NULL ORDER BY repo, issue_num` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 806 |
| `SELECT repo, issue_num, started_at, ended_at FROM sessions` | `(*SQLiteStore).totalTimes` | `internal/store/sqlite.go` | 848 |
| `SELECT started_at, ended_at FROM sessions WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).totalTime` | `internal/store/sqlite.go` | 893 |
| `INSERT INTO issue_budgets (repo, issue_num, budget) VALUES (?, ?, ?) 		 ON CONFLICT(repo, issue_num) DO UPDATE SET budget = excluded.budget` | `(*SQLiteStore).SetBudget` | `internal/store/sqlite.go` | 938 |
| `DELETE FROM issue_budgets WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).UnsetBudget` | `internal/store/sqlite.go` | 951 |
| `SELECT repo, issue_num, budget FROM issue_budgets ORDER BY repo, issue_num` | `(*SQLiteStore).ListBudgets` | `internal/store/sqlite.go` | 963 |
| `SELECT budget FROM issue_budgets WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).GetBudget` | `internal/store/sqlite.go` | 991 |
| ` 		SELECT COALESCE(ue.session_id, ''), ue.repo, ue.model, COUNT(*), 		       SUM(ue.input_fresh), SUM(ue.cache_write), SUM(ue.cache_read), 		       SUM(ue.output), SUM(ue.reasoning), 		       MIN(ue.at), MAX(ue.at) 		FROM usage_entries ue 		WHERE ue.issue_num IS NULL` | `(*SQLiteStore).ListUnattributed` | `internal/store/sqlite.go` | 1028 |
| `UPDATE usage_entries SET issue_num = ? WHERE session_id = ?` | `(*SQLiteStore).setSessionIssue` | `internal/store/sqlite.go` | 1116 |
| `UPDATE sessions SET issue_num = ? WHERE id = ?` | `(*SQLiteStore).setSessionIssue` | `internal/store/sqlite.go` | 1127 |

