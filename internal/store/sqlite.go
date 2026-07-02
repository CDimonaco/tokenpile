package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite" // registers the sqlite3 driver

	"github.com/cdimonaco/tokenpile/internal/usage"
)

var ErrSessionNotFound = errors.New("session not found")

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

CREATE INDEX IF NOT EXISTS idx_usage_entries_repo_issue ON usage_entries (repo, issue_num);
CREATE INDEX IF NOT EXISTS idx_usage_entries_at ON usage_entries (at);
CREATE INDEX IF NOT EXISTS idx_sessions_repo_issue ON sessions (repo, issue_num);
`

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

	return &SQLiteStore{db: db, pricing: pricer}, nil
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

func (s *SQLiteStore) ListSessions(ctx context.Context, repo string, issueNum int) ([]usage.Session, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, repo, issue_num, started_at, ended_at FROM sessions WHERE repo = ? AND issue_num = ? ORDER BY started_at`,
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
		var endedAt sql.NullString

		if err = rows.Scan(&sess.ID, &sess.Repo, &sess.IssueNum, &startedAt, &endedAt); err != nil {
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

		sessions = append(sessions, sess)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}

	return sessions, nil
}

func (s *SQLiteStore) ListIssues(ctx context.Context, filter usage.Filter) ([]usage.TrackedIssue, error) {
	query := `
		SELECT repo, issue_num, model, SUM(tokens_in), SUM(tokens_out)
		FROM usage_entries
		WHERE 1=1`
	args := []any{}

	if filter.Repo != "" {
		query += " AND repo = ?"
		args = append(args, filter.Repo)
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

	query += " GROUP BY repo, issue_num, model ORDER BY repo, issue_num"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}
	defer rows.Close()

	type issueKey struct {
		repo     string
		issueNum int
	}
	issueMap := make(map[issueKey]*usage.TrackedIssue)
	var issueOrder []issueKey

	for rows.Next() {
		var repo, model string
		var issueNum, tokensIn, tokensOut int

		if err = rows.Scan(&repo, &issueNum, &model, &tokensIn, &tokensOut); err != nil {
			return nil, fmt.Errorf("scan issue: %w", err)
		}

		k := issueKey{repo, issueNum}
		if _, ok := issueMap[k]; !ok {
			issueMap[k] = &usage.TrackedIssue{Repo: repo, IssueNum: issueNum}
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

	issues := make([]usage.TrackedIssue, 0, len(issueOrder))
	for _, k := range issueOrder {
		ti := issueMap[k]
		ti.TotalTime = s.totalTime(ctx, k.repo, k.issueNum)
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
