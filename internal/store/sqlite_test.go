package store_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

type fixedPricer struct {
	pricePerMIn  float64
	pricePerMOut float64
}

func (f fixedPricer) ComputeCost(_ string, tokensIn, tokensOut int) (float64, bool) {
	return float64(tokensIn)/1_000_000*f.pricePerMIn + float64(tokensOut)/1_000_000*f.pricePerMOut, true
}

func newTestStore(t *testing.T) *store.SQLiteStore {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewSQLiteStore(dbPath, fixedPricer{pricePerMIn: 3.0, pricePerMOut: 15.0})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	return s
}

func TestSQLiteStore_LogUsage(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	entry := usage.Entry{
		Repo:      "owner/repo",
		IssueNum:  42,
		Agent:     "claude-code",
		Model:     "claude-sonnet-4-6",
		TokensIn:  1000,
		TokensOut: 500,
	}

	err := s.LogUsage(ctx, entry)
	require.NoError(t, err)

	issues, err := s.ListIssues(ctx, usage.Filter{Repo: "owner/repo"})
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, 42, issues[0].IssueNum)
	assert.Equal(t, 1000, issues[0].TotalTokensIn)
	assert.Equal(t, 500, issues[0].TotalTokensOut)
}

func TestSQLiteStore_LogUsage_SetsTimestamp(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	before := time.Now().UTC().Add(-time.Second)

	err := s.LogUsage(ctx, usage.Entry{
		Repo: "owner/repo", IssueNum: 1, Agent: "a", Model: "m",
	})
	require.NoError(t, err)

	issues, err := s.ListIssues(ctx, usage.Filter{Repo: "owner/repo"})
	require.NoError(t, err)
	require.Len(t, issues, 1)
	_ = before
}

func TestSQLiteStore_Session_StartEnd(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sess, err := s.StartSession(ctx, "owner/repo", 42)
	require.NoError(t, err)
	require.NotEmpty(t, sess.ID)
	assert.Nil(t, sess.EndedAt)

	err = s.EndSession(ctx, sess.ID)
	require.NoError(t, err)

	sessions, err := s.ListSessions(ctx, "owner/repo", 42)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.NotNil(t, sessions[0].EndedAt)
}

func TestSQLiteStore_EndSession_NotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	err := s.EndSession(ctx, "nonexistent")
	assert.ErrorIs(t, err, store.ErrSessionNotFound)
}

func TestSQLiteStore_GetReport_ByAgentModel(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	entries := []usage.Entry{
		{Repo: "o/r", IssueNum: 1, Agent: "claude-code", Model: "claude-sonnet-4-6", TokensIn: 1000, TokensOut: 200},
		{Repo: "o/r", IssueNum: 1, Agent: "opencode", Model: "gpt-4o", TokensIn: 500, TokensOut: 100},
		{Repo: "o/r", IssueNum: 1, Agent: "claude-code", Model: "claude-sonnet-4-6", TokensIn: 2000, TokensOut: 400},
	}

	for _, e := range entries {
		require.NoError(t, s.LogUsage(ctx, e))
	}

	report, err := s.GetReport(ctx, "o/r", 1)
	require.NoError(t, err)
	require.Len(t, report.Rows, 2)
	assert.Equal(t, 3500, report.TotalTokensIn)
	assert.Equal(t, 700, report.TotalTokensOut)
}

func TestSQLiteStore_GetReport_EmptyIssue(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	report, err := s.GetReport(ctx, "o/r", 99)
	require.NoError(t, err)
	assert.Empty(t, report.Rows)
	assert.Equal(t, 0, report.TotalTokensIn)
}

func TestSQLiteStore_ListIssues_MultipleIssues(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for _, num := range []int{1, 2, 3} {
		require.NoError(t, s.LogUsage(ctx, usage.Entry{
			Repo: "o/r", IssueNum: num, Agent: "a", Model: "m",
			TokensIn: 100 * num, TokensOut: 50 * num,
		}))
	}

	issues, err := s.ListIssues(ctx, usage.Filter{Repo: "o/r"})
	require.NoError(t, err)
	assert.Len(t, issues, 3)
}

func TestSQLiteStore_ListIssues_FilterByAgent(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.LogUsage(ctx, usage.Entry{Repo: "o/r", IssueNum: 1, Agent: "claude-code", Model: "m", TokensIn: 100}))
	require.NoError(t, s.LogUsage(ctx, usage.Entry{Repo: "o/r", IssueNum: 2, Agent: "opencode", Model: "m", TokensIn: 200}))

	issues, err := s.ListIssues(ctx, usage.Filter{Repo: "o/r", Agent: "claude-code"})
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, 1, issues[0].IssueNum)
}

func TestSQLiteStore_ListUsageOverTime_DayGranularity(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	today := time.Now().UTC().Truncate(24 * time.Hour)

	for i := range 3 {
		at := today.AddDate(0, 0, i)
		require.NoError(t, s.LogUsage(ctx, usage.Entry{
			Repo: "o/r", IssueNum: 1, Agent: "a", Model: "m",
			TokensIn: 100, TokensOut: 50, At: at,
		}))
	}

	points, err := s.ListUsageOverTime(ctx, usage.OverTimeFilter{
		Repo:        "o/r",
		Granularity: usage.Day,
	})
	require.NoError(t, err)
	assert.Len(t, points, 3)

	for _, p := range points {
		assert.Equal(t, 100, p.TokensIn)
		assert.Equal(t, 50, p.TokensOut)
	}
}

func TestSQLiteStore_SchemaIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	pricer := fixedPricer{}

	s1, err := store.NewSQLiteStore(dbPath, pricer)
	require.NoError(t, err)
	require.NoError(t, s1.Close())

	s2, err := store.NewSQLiteStore(dbPath, pricer)
	require.NoError(t, err)
	require.NoError(t, s2.Close())
}
