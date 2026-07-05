package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite" // registers the sqlite3 driver

	"github.com/cdimonaco/tokenpile/internal/usage"
)

var (
	ErrSessionNotFound    = errors.New("session not found")
	ErrIssueCacheNotFound = errors.New("issue not in cache")
	ErrBudgetNotFound     = errors.New("budget not set")
)

const schema = `
CREATE TABLE IF NOT EXISTS usage_entries (
    id          TEXT PRIMARY KEY,
    repo        TEXT NOT NULL,
    issue_num   INTEGER NOT NULL,
    agent       TEXT NOT NULL,
    model       TEXT NOT NULL,
    tokens_in   INTEGER NOT NULL,
    tokens_out  INTEGER NOT NULL,
    session_id  TEXT,
    at          TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id          TEXT PRIMARY KEY,
    repo        TEXT NOT NULL,
    issue_num   INTEGER NOT NULL,
    started_at  TEXT NOT NULL,
    ended_at    TEXT
);

CREATE TABLE IF NOT EXISTS issue_cache (
    repo        TEXT NOT NULL,
    issue_num   INTEGER NOT NULL,
    title       TEXT NOT NULL DEFAULT '',
    labels      TEXT NOT NULL DEFAULT '[]',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    PRIMARY KEY (repo, issue_num)
);

CREATE TABLE IF NOT EXISTS issue_budgets (
    repo        TEXT NOT NULL,
    issue_num   INTEGER NOT NULL,
    budget      REAL NOT NULL,
    PRIMARY KEY (repo, issue_num)
);

CREATE INDEX IF NOT EXISTS idx_usage_entries_repo_issue ON usage_entries (repo, issue_num);
CREATE INDEX IF NOT EXISTS idx_usage_entries_at ON usage_entries (at);
CREATE INDEX IF NOT EXISTS idx_sessions_repo_issue ON sessions (repo, issue_num);
`

// migrations holds ALTER TABLE statements for additive schema changes.
// Each statement is run after the base schema is applied. SQLite returns
// "duplicate column name" when the column already exists; that error is
// silently ignored so the function is safe to run on every startup.
var migrations = []string{
	`ALTER TABLE sessions ADD COLUMN note TEXT`,
	`ALTER TABLE sessions ADD COLUMN tags TEXT`,
	`ALTER TABLE sessions ADD COLUMN last_activity_at TEXT`,
}

type SQLiteStore struct {
	db      *sql.DB
	pricing Pricer
}

type Pricer interface {
	ComputeCost(model string, tokensIn, tokensOut int) (float64, bool)
}

func NewSQLiteStore(dbPath string, pricer Pricer) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	if _, err = db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	if err = runMigrations(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &SQLiteStore{db: db, pricing: pricer}, nil
}

func runMigrations(db *sql.DB) error {
	for _, stmt := range migrations {
		if _, err := db.Exec(stmt); err != nil {
			// SQLite returns "duplicate column name" when the column already exists.
			if strings.Contains(err.Error(), "duplicate column name") {
				continue
			}

			return fmt.Errorf("migration %q: %w", stmt, err)
		}
	}

	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) LogUsage(ctx context.Context, entry usage.Entry) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	if entry.At.IsZero() {
		entry.At = time.Now().UTC()
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO usage_entries (id, repo, issue_num, agent, model, tokens_in, tokens_out, session_id, at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.Repo, entry.IssueNum, entry.Agent, entry.Model,
		entry.TokensIn, entry.TokensOut, entry.SessionID, entry.At.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert usage entry: %w", err)
	}

	return nil
}

func (s *SQLiteStore) UpsertIssueCache(ctx context.Context, cache *usage.IssueCache) error {
	labels := cache.Labels
	if labels == nil {
		labels = []string{}
	}

	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO issue_cache (repo, issue_num, title, labels, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(repo, issue_num) DO UPDATE SET
			title = excluded.title,
			labels = excluded.labels,
			updated_at = excluded.updated_at`,
		cache.Repo, cache.IssueNum, cache.Title, string(labelsJSON), now, now,
	)
	if err != nil {
		return fmt.Errorf("upsert issue cache: %w", err)
	}

	return nil
}

func (s *SQLiteStore) GetIssueCache(ctx context.Context, repo string, issueNum int) (*usage.IssueCache, error) {
	var cache usage.IssueCache
	var labelsJSON, createdAt, updatedAt string

	err := s.db.QueryRowContext(ctx,
		`SELECT repo, issue_num, title, labels, created_at, updated_at
		 FROM issue_cache WHERE repo = ? AND issue_num = ?`,
		repo, issueNum,
	).Scan(&cache.Repo, &cache.IssueNum, &cache.Title, &labelsJSON, &createdAt, &updatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrIssueCacheNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get issue cache: %w", err)
	}

	if jsonErr := json.Unmarshal([]byte(labelsJSON), &cache.Labels); jsonErr != nil {
		cache.Labels = nil
	}

	cache.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	cache.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &cache, nil
}

func (s *SQLiteStore) ListEntries(ctx context.Context, filter usage.Filter) ([]usage.Entry, error) {
	query := `SELECT ue.id, ue.repo, ue.issue_num, ue.agent, ue.model,
		ue.tokens_in, ue.tokens_out, ue.session_id, ue.at,
		COALESCE(ic.title, ''), COALESCE(ic.labels, '[]')
		FROM usage_entries ue
		LEFT JOIN issue_cache ic ON ic.repo = ue.repo AND ic.issue_num = ue.issue_num
		WHERE 1=1`
	args := []any{}

	if filter.Repo != "" {
		query += " AND ue.repo = ?"
		args = append(args, filter.Repo)
	}

	if filter.IssueNum != 0 {
		query += " AND ue.issue_num = ?"
		args = append(args, filter.IssueNum)
	}

	if filter.Agent != "" {
		query += " AND ue.agent = ?"
		args = append(args, filter.Agent)
	}

	if filter.Model != "" {
		query += " AND ue.model = ?"
		args = append(args, filter.Model)
	}

	if filter.From != nil {
		query += " AND ue.at >= ?"
		args = append(args, filter.From.UTC().Format(time.RFC3339))
	}

	if filter.To != nil {
		query += " AND ue.at <= ?"
		args = append(args, filter.To.UTC().Format(time.RFC3339))
	}

	query += " ORDER BY ue.at"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	defer rows.Close()

	var entries []usage.Entry

	for rows.Next() {
		var e usage.Entry
		var atStr, labelsJSON string
		var sessionID sql.NullString

		if err = rows.Scan(&e.ID, &e.Repo, &e.IssueNum, &e.Agent, &e.Model,
			&e.TokensIn, &e.TokensOut, &sessionID, &atStr, &e.IssueTitle, &labelsJSON); err != nil {
			return nil, fmt.Errorf("scan entry: %w", err)
		}

		e.At, err = time.Parse(time.RFC3339, atStr)
		if err != nil {
			return nil, fmt.Errorf("parse entry at: %w", err)
		}

		if sessionID.Valid {
			e.SessionID = sessionID.String
		}

		if jsonErr := json.Unmarshal([]byte(labelsJSON), &e.IssueLabels); jsonErr != nil {
			e.IssueLabels = nil
		}

		entries = append(entries, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entries: %w", err)
	}

	return entries, nil
}

func (s *SQLiteStore) StartSession(ctx context.Context, repo string, issueNum int) (*usage.Session, error) {
	sess := usage.Session{
		ID:        uuid.New().String(),
		Repo:      repo,
		IssueNum:  issueNum,
		StartedAt: time.Now().UTC(),
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, repo, issue_num, started_at) VALUES (?, ?, ?, ?)`,
		sess.ID, sess.Repo, sess.IssueNum, sess.StartedAt.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}

	return &sess, nil
}

func (s *SQLiteStore) EndSession(ctx context.Context, sessionID string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	res, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET ended_at = ? WHERE id = ? AND ended_at IS NULL`,
		now, sessionID,
	)
	if err != nil {
		return fmt.Errorf("end session: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("end session rows affected: %w", err)
	}

	if n == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (s *SQLiteStore) UpdateSessionActivity(ctx context.Context, sessionID string, at time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET last_activity_at = ? WHERE id = ? AND ended_at IS NULL`,
		at.UTC().Format(time.RFC3339), sessionID,
	)
	if err != nil {
		return fmt.Errorf("update session activity: %w", err)
	}

	return nil
}

func (s *SQLiteStore) EndSessionAt(ctx context.Context, sessionID string, at time.Time) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET ended_at = ? WHERE id = ? AND ended_at IS NULL`,
		at.UTC().Format(time.RFC3339), sessionID,
	)
	if err != nil {
		return fmt.Errorf("end session: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("end session rows affected: %w", err)
	}

	if n == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (s *SQLiteStore) ListSessions(ctx context.Context, repo string, issueNum int) ([]usage.Session, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, repo, issue_num, started_at, ended_at, last_activity_at, note, tags
		 FROM sessions WHERE repo = ? AND issue_num = ? ORDER BY started_at`,
		repo,
		issueNum,
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []usage.Session

	for rows.Next() {
		var sess usage.Session
		var startedAt string
		var endedAt, lastActivityAt, note, tagsJSON sql.NullString

		if err = rows.Scan(
			&sess.ID, &sess.Repo, &sess.IssueNum,
			&startedAt, &endedAt, &lastActivityAt,
			&note, &tagsJSON,
		); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}

		sess.StartedAt, err = time.Parse(time.RFC3339, startedAt)
		if err != nil {
			return nil, fmt.Errorf("parse session started_at: %w", err)
		}

		if endedAt.Valid {
			t, parseErr := time.Parse(time.RFC3339, endedAt.String)
			if parseErr != nil {
				return nil, fmt.Errorf("parse session ended_at: %w", parseErr)
			}

			sess.EndedAt = &t
		}

		if lastActivityAt.Valid && lastActivityAt.String != "" {
			t, parseErr := time.Parse(time.RFC3339, lastActivityAt.String)
			if parseErr == nil {
				sess.LastActivityAt = t
			}
		}

		if sess.LastActivityAt.IsZero() {
			sess.LastActivityAt = sess.StartedAt
		}

		if note.Valid {
			sess.Note = note.String
		}

		if tagsJSON.Valid && tagsJSON.String != "" {
			if jsonErr := json.Unmarshal([]byte(tagsJSON.String), &sess.Tags); jsonErr != nil {
				sess.Tags = []string{}
			}
		}

		if sess.Tags == nil {
			sess.Tags = []string{}
		}

		sessions = append(sessions, sess)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}

	return sessions, nil
}

func (s *SQLiteStore) UpdateSessionAnnotations(
	ctx context.Context,
	sessionID string,
	note *string,
	tags []string,
) error {
	// Fetch existing tags to merge.
	var existingTagsJSON sql.NullString

	err := s.db.QueryRowContext(ctx, `SELECT tags FROM sessions WHERE id = ?`, sessionID).Scan(&existingTagsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSessionNotFound
	}

	if err != nil {
		return fmt.Errorf("fetch session tags: %w", err)
	}

	// Merge tags: union of existing + new, no duplicates.
	tagSet := make(map[string]struct{})

	if existingTagsJSON.Valid && existingTagsJSON.String != "" {
		var existing []string
		if jsonErr := json.Unmarshal([]byte(existingTagsJSON.String), &existing); jsonErr == nil {
			for _, t := range existing {
				tagSet[t] = struct{}{}
			}
		}
	}

	for _, t := range tags {
		if t != "" {
			tagSet[t] = struct{}{}
		}
	}

	merged := make([]string, 0, len(tagSet))
	for t := range tagSet {
		merged = append(merged, t)
	}

	tagsJSON, err := json.Marshal(merged)
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	if note != nil {
		_, err = s.db.ExecContext(ctx,
			`UPDATE sessions SET note = ?, tags = ? WHERE id = ?`,
			*note, string(tagsJSON), sessionID,
		)
	} else {
		_, err = s.db.ExecContext(ctx,
			`UPDATE sessions SET tags = ? WHERE id = ?`,
			string(tagsJSON), sessionID,
		)
	}

	if err != nil {
		return fmt.Errorf("update session annotations: %w", err)
	}

	return nil
}

func (s *SQLiteStore) ListIssues(ctx context.Context, filter usage.Filter) ([]usage.TrackedIssue, error) {
	query := `
		SELECT ue.repo, ue.issue_num, ue.model, SUM(ue.tokens_in), SUM(ue.tokens_out),
		       COALESCE(ic.title, ''), COALESCE(ic.labels, '[]'), ib.budget
		FROM usage_entries ue
		LEFT JOIN issue_cache ic ON ic.repo = ue.repo AND ic.issue_num = ue.issue_num
		LEFT JOIN issue_budgets ib ON ib.repo = ue.repo AND ib.issue_num = ue.issue_num
		WHERE 1=1`
	args := []any{}

	if filter.Repo != "" {
		query += " AND ue.repo = ?"
		args = append(args, filter.Repo)
	}

	if filter.Agent != "" {
		query += " AND ue.agent = ?"
		args = append(args, filter.Agent)
	}

	if filter.Model != "" {
		query += " AND ue.model = ?"
		args = append(args, filter.Model)
	}

	if filter.From != nil {
		query += " AND ue.at >= ?"
		args = append(args, filter.From.UTC().Format(time.RFC3339))
	}

	if filter.To != nil {
		query += " AND ue.at <= ?"
		args = append(args, filter.To.UTC().Format(time.RFC3339))
	}

	query += " GROUP BY ue.repo, ue.issue_num, ue.model, ic.title, ic.labels, ib.budget ORDER BY ue.repo, ue.issue_num"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}
	defer rows.Close()

	issueMap := make(map[issueKey]*usage.TrackedIssue)
	var issueOrder []issueKey

	for rows.Next() {
		var repo, model, title, labelsJSON string
		var issueNum, tokensIn, tokensOut int
		var budget sql.NullFloat64

		if err = rows.Scan(&repo, &issueNum, &model, &tokensIn, &tokensOut, &title, &labelsJSON, &budget); err != nil {
			return nil, fmt.Errorf("scan issue: %w", err)
		}

		k := issueKey{repo, issueNum}
		if _, ok := issueMap[k]; !ok {
			var labels []string
			if jsonErr := json.Unmarshal([]byte(labelsJSON), &labels); jsonErr != nil {
				labels = nil
			}

			ti := &usage.TrackedIssue{Repo: repo, IssueNum: issueNum, Title: title, Labels: labels}
			if budget.Valid {
				v := budget.Float64
				ti.Budget = &v
			}

			issueMap[k] = ti
			issueOrder = append(issueOrder, k)
		}

		cost, _ := s.pricing.ComputeCost(model, tokensIn, tokensOut)
		issueMap[k].TotalTokensIn += tokensIn
		issueMap[k].TotalTokensOut += tokensOut
		issueMap[k].TotalCost += cost
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issues: %w", err)
	}

	totals := s.totalTimes(ctx, filter.Repo)

	issues := make([]usage.TrackedIssue, 0, len(issueOrder))
	for _, k := range issueOrder {
		ti := issueMap[k]
		ti.TotalTime = totals[k]
		issues = append(issues, *ti)
	}

	return issues, nil
}

func (s *SQLiteStore) GetReport(ctx context.Context, repo string, issueNum int) (*usage.Report, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT agent, model, COUNT(*), SUM(tokens_in), SUM(tokens_out)
		 FROM usage_entries
		 WHERE repo = ? AND issue_num = ?
		 GROUP BY agent, model
		 ORDER BY agent, model`,
		repo, issueNum,
	)
	if err != nil {
		return nil, fmt.Errorf("get report: %w", err)
	}
	defer rows.Close()

	report := &usage.Report{
		IssueNum: issueNum,
		Repo:     repo,
	}

	for rows.Next() {
		var row usage.ReportRow
		if err = rows.Scan(&row.Agent, &row.Model, &row.Calls, &row.TokensIn, &row.TokensOut); err != nil {
			return nil, fmt.Errorf("scan report row: %w", err)
		}

		cost, _ := s.pricing.ComputeCost(row.Model, row.TokensIn, row.TokensOut)
		row.Cost = cost

		report.Rows = append(report.Rows, row)
		report.TotalTokensIn += row.TokensIn
		report.TotalTokensOut += row.TokensOut
		report.TotalCost += cost
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate report rows: %w", err)
	}

	report.TotalTime = s.totalTime(ctx, repo, issueNum)

	return report, nil
}

func (s *SQLiteStore) ListUsageOverTime(ctx context.Context, filter usage.OverTimeFilter) ([]usage.Point, error) {
	// For week granularity, compute the Monday of each week as YYYY-MM-DD so
	// we can always parse with the same "2006-01-02" layout. SQLite's %Y-W%W
	// produces strings like "2026-W26" which Go's time package cannot parse.
	periodExpr := "strftime('%Y-%m-%d', at)"
	if filter.Granularity == usage.Week {
		periodExpr = "date(at, '-' || cast((cast(strftime('%w', at) as integer) + 6) % 7 as text) || ' days')"
	}

	query := fmt.Sprintf(`
		SELECT %s as period, model, SUM(tokens_in), SUM(tokens_out)
		FROM usage_entries
		WHERE 1=1`, periodExpr)
	args := []any{}

	if filter.Repo != "" {
		query += " AND repo = ?"
		args = append(args, filter.Repo)
	}

	if filter.IssueNum != nil {
		query += " AND issue_num = ?"
		args = append(args, *filter.IssueNum)
	}

	if filter.Agent != "" {
		query += " AND agent = ?"
		args = append(args, filter.Agent)
	}

	if filter.Model != "" {
		query += " AND model = ?"
		args = append(args, filter.Model)
	}

	if filter.From != nil {
		query += " AND at >= ?"
		args = append(args, filter.From.UTC().Format(time.RFC3339))
	}

	if filter.To != nil {
		query += " AND at <= ?"
		args = append(args, filter.To.UTC().Format(time.RFC3339))
	}

	query += " GROUP BY period, model ORDER BY period"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list usage over time: %w", err)
	}
	defer rows.Close()

	// Aggregate rows by period, computing cost per model then summing.
	type periodKey = string
	type periodData struct {
		date      time.Time
		tokensIn  int
		tokensOut int
		cost      float64
	}

	periodMap := make(map[periodKey]*periodData)
	var periodOrder []string

	for rows.Next() {
		var period, model string
		var tokensIn, tokensOut int

		if err = rows.Scan(&period, &model, &tokensIn, &tokensOut); err != nil {
			return nil, fmt.Errorf("scan usage point: %w", err)
		}

		if _, exists := periodMap[period]; !exists {
			t, parseErr := time.Parse("2006-01-02", period)
			if parseErr != nil {
				t = time.Time{}
			}

			periodMap[period] = &periodData{date: t}
			periodOrder = append(periodOrder, period)
		}

		cost, _ := s.pricing.ComputeCost(model, tokensIn, tokensOut)
		periodMap[period].tokensIn += tokensIn
		periodMap[period].tokensOut += tokensOut
		periodMap[period].cost += cost
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate usage points: %w", err)
	}

	points := make([]usage.Point, 0, len(periodOrder))
	for _, period := range periodOrder {
		d := periodMap[period]
		points = append(points, usage.Point{
			Date:      d.date,
			TokensIn:  d.tokensIn,
			TokensOut: d.tokensOut,
			Cost:      d.cost,
		})
	}

	return points, nil
}

func (s *SQLiteStore) ListTrackedIssueRefs(ctx context.Context) ([]usage.TrackedIssueRef, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT repo, issue_num FROM usage_entries ORDER BY repo, issue_num`,
	)
	if err != nil {
		return nil, fmt.Errorf("list tracked issue refs: %w", err)
	}
	defer rows.Close()

	var refs []usage.TrackedIssueRef

	for rows.Next() {
		var ref usage.TrackedIssueRef
		if err = rows.Scan(&ref.Repo, &ref.IssueNum); err != nil {
			return nil, fmt.Errorf("scan tracked issue ref: %w", err)
		}

		refs = append(refs, ref)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tracked issue refs: %w", err)
	}

	return refs, nil
}

type issueKey struct {
	repo     string
	issueNum int
}

// totalTimes aggregates session durations per issue in a single query,
// optionally filtered by repo. Open sessions count up to now.
func (s *SQLiteStore) totalTimes(ctx context.Context, repo string) map[issueKey]time.Duration {
	query := `SELECT repo, issue_num, started_at, ended_at FROM sessions`
	args := []any{}

	if repo != "" {
		query += ` WHERE repo = ?`
		args = append(args, repo)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	totals := make(map[issueKey]time.Duration)
	now := time.Now().UTC()

	for rows.Next() {
		var k issueKey
		var startedAtStr string
		var endedAtNull sql.NullString

		if err = rows.Scan(&k.repo, &k.issueNum, &startedAtStr, &endedAtNull); err != nil {
			continue
		}

		startedAt, parseErr := time.Parse(time.RFC3339, startedAtStr)
		if parseErr != nil {
			continue
		}

		end := now

		if endedAtNull.Valid {
			endedAt, parseEndErr := time.Parse(time.RFC3339, endedAtNull.String)
			if parseEndErr != nil {
				continue
			}

			end = endedAt
		}

		totals[k] += end.Sub(startedAt)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		slog.Error("iterate sessions for total times", "err", rowsErr)
	}

	return totals
}

func (s *SQLiteStore) totalTime(ctx context.Context, repo string, issueNum int) time.Duration {
	rows, err := s.db.QueryContext(ctx,
		`SELECT started_at, ended_at FROM sessions WHERE repo = ? AND issue_num = ?`,
		repo, issueNum,
	)
	if err != nil {
		return 0
	}
	defer rows.Close()

	var total time.Duration
	now := time.Now().UTC()

	for rows.Next() {
		var startedAtStr string
		var endedAtNull sql.NullString

		if err = rows.Scan(&startedAtStr, &endedAtNull); err != nil {
			continue
		}

		startedAt, parseErr := time.Parse(time.RFC3339, startedAtStr)
		if parseErr != nil {
			continue
		}

		if endedAtNull.Valid {
			endedAt, parseEndErr := time.Parse(time.RFC3339, endedAtNull.String)
			if parseEndErr != nil {
				continue
			}

			total += endedAt.Sub(startedAt)
		} else {
			total += now.Sub(startedAt)
		}
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		slog.Error("iterate sessions for total time", "err", rowsErr)
	}

	return total
}

func (s *SQLiteStore) SetBudget(ctx context.Context, repo string, issueNum int, budget float64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO issue_budgets (repo, issue_num, budget) VALUES (?, ?, ?)
		 ON CONFLICT(repo, issue_num) DO UPDATE SET budget = excluded.budget`,
		repo, issueNum, budget,
	)
	if err != nil {
		return fmt.Errorf("set budget: %w", err)
	}

	return nil
}

func (s *SQLiteStore) UnsetBudget(ctx context.Context, repo string, issueNum int) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM issue_budgets WHERE repo = ? AND issue_num = ?`,
		repo, issueNum,
	)
	if err != nil {
		return fmt.Errorf("unset budget: %w", err)
	}

	return nil
}

func (s *SQLiteStore) GetBudget(ctx context.Context, repo string, issueNum int) (*float64, error) {
	var budget float64
	err := s.db.QueryRowContext(ctx,
		`SELECT budget FROM issue_budgets WHERE repo = ? AND issue_num = ?`,
		repo, issueNum,
	).Scan(&budget)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBudgetNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get budget: %w", err)
	}

	return &budget, nil
}
