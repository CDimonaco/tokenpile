# SQL Queries

**Root:** `/Users/cdimonaco/code/github.com/cdimonaco/tokenpile`  
**Generated:** 2026-07-05 15:20:00 UTC  

---

| Query | Function | File | Line |
|-------|----------|------|------|
| `INSERT INTO usage_entries (id, repo, issue_num, agent, model, tokens_in, tokens_out, session_id, at) 		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)` | `(*SQLiteStore).LogUsage` | `internal/store/sqlite.go` | 136 |
| ` 		INSERT INTO issue_cache (repo, issue_num, title, labels, created_at, updated_at) 		VALUES (?, ?, ?, ?, ?, ?) 		ON CONFLICT(repo, issue_num) DO UPDATE SET 			title = excluded.title, 			labels = excluded.labels, 			updated_at = excluded.updated_at` | `(*SQLiteStore).UpsertIssueCache` | `internal/store/sqlite.go` | 162 |
| `SELECT repo, issue_num, title, labels, created_at, updated_at 		 FROM issue_cache WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).GetIssueCache` | `internal/store/sqlite.go` | 182 |
| `SELECT ue.id, ue.repo, ue.issue_num, ue.agent, ue.model, 		ue.tokens_in, ue.tokens_out, ue.session_id, ue.at, 		COALESCE(ic.title, ''), COALESCE(ic.labels, '[]') 		FROM usage_entries ue 		LEFT JOIN issue_cache ic ON ic.repo = ue.repo AND ic.issue_num = ue.issue_num 		WHERE 1=1` | `(*SQLiteStore).ListEntries` | `internal/store/sqlite.go` | 247 |
| `INSERT INTO sessions (id, repo, issue_num, started_at) VALUES (?, ?, ?, ?)` | `(*SQLiteStore).StartSession` | `internal/store/sqlite.go` | 296 |
| `UPDATE sessions SET ended_at = ? WHERE id = ? AND ended_at IS NULL` | `(*SQLiteStore).EndSession` | `internal/store/sqlite.go` | 310 |
| `UPDATE sessions SET last_activity_at = ? WHERE id = ? AND ended_at IS NULL` | `(*SQLiteStore).UpdateSessionActivity` | `internal/store/sqlite.go` | 331 |
| `UPDATE sessions SET ended_at = ? WHERE id = ? AND ended_at IS NULL` | `(*SQLiteStore).EndSessionAt` | `internal/store/sqlite.go` | 343 |
| `SELECT id, repo, issue_num, started_at, ended_at, last_activity_at, note, tags 		 FROM sessions WHERE repo = ? AND issue_num = ? ORDER BY started_at` | `(*SQLiteStore).ListSessions` | `internal/store/sqlite.go` | 364 |
| `SELECT id, repo, issue_num, started_at, ended_at, last_activity_at, note, tags 		 FROM sessions ORDER BY repo, issue_num, started_at` | `(*SQLiteStore).ListAllSessions` | `internal/store/sqlite.go` | 380 |
| `SELECT tags FROM sessions WHERE id = ?` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 468 |
| `UPDATE sessions SET note = ?, tags = ? WHERE id = ?` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 506 |
| `UPDATE sessions SET tags = ? WHERE id = ?` | `(*SQLiteStore).UpdateSessionAnnotations` | `internal/store/sqlite.go` | 511 |
| ` 		SELECT ue.repo, ue.issue_num, ue.model, SUM(ue.tokens_in), SUM(ue.tokens_out), 		       COALESCE(ic.title, ''), COALESCE(ic.labels, '[]'), ib.budget 		FROM usage_entries ue 		LEFT JOIN issue_cache ic ON ic.repo = ue.repo AND ic.issue_num = ue.issue_num 		LEFT JOIN issue_budgets ib ON ib.repo = ue.repo AND ib.issue_num = ue.issue_num 		WHERE 1=1` | `(*SQLiteStore).ListIssues` | `internal/store/sqlite.go` | 561 |
| `SELECT agent, model, COUNT(*), SUM(tokens_in), SUM(tokens_out) 		 FROM usage_entries 		 WHERE repo = ? AND issue_num = ? 		 GROUP BY agent, model 		 ORDER BY agent, model` | `(*SQLiteStore).GetReport` | `internal/store/sqlite.go` | 619 |
| `SELECT DISTINCT repo, issue_num FROM usage_entries ORDER BY repo, issue_num` | `(*SQLiteStore).ListTrackedIssueRefs` | `internal/store/sqlite.go` | 769 |
| `SELECT repo, issue_num, started_at, ended_at FROM sessions` | `(*SQLiteStore).totalTimes` | `internal/store/sqlite.go` | 811 |
| `SELECT started_at, ended_at FROM sessions WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).totalTime` | `internal/store/sqlite.go` | 856 |
| `INSERT INTO issue_budgets (repo, issue_num, budget) VALUES (?, ?, ?) 		 ON CONFLICT(repo, issue_num) DO UPDATE SET budget = excluded.budget` | `(*SQLiteStore).SetBudget` | `internal/store/sqlite.go` | 901 |
| `DELETE FROM issue_budgets WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).UnsetBudget` | `internal/store/sqlite.go` | 914 |
| `SELECT repo, issue_num, budget FROM issue_budgets ORDER BY repo, issue_num` | `(*SQLiteStore).ListBudgets` | `internal/store/sqlite.go` | 926 |
| `SELECT budget FROM issue_budgets WHERE repo = ? AND issue_num = ?` | `(*SQLiteStore).GetBudget` | `internal/store/sqlite.go` | 954 |

