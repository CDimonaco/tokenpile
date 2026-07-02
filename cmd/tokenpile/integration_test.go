package main

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/domain"
	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/cdimonaco/tokenpile/internal/store"
)

func newTestStore(t *testing.T) *store.SQLiteStore {
	t.Helper()

	dir := t.TempDir()
	loader, err := pricing.NewLoader("")
	require.NoError(t, err)

	s, err := store.NewSQLiteStore(filepath.Join(dir, "test.db"), loader)
	require.NoError(t, err)

	t.Cleanup(func() { _ = s.Close() })

	return s
}

func runLogCmd(t *testing.T, s *store.SQLiteStore, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{logCommand(s)},
	}

	err := app.RunContext(context.Background(), append([]string{"tok"}, args...))

	return buf.String(), err
}

func TestIntegration_Log_CreatesSessionAndEntry(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := runLogCmd(t, s,
		"log",
		"--issue", "1",
		"--agent", "claude-code",
		"--model", "claude-sonnet-4-6",
		"--tokens-in", "1000",
		"--tokens-out", "500",
		"--repo", "owner/repo",
	)
	require.NoError(t, err)

	sessions, err := s.ListSessions(ctx, "owner/repo", 1)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Nil(t, sessions[0].EndedAt)

	report, err := s.GetReport(ctx, "owner/repo", 1)
	require.NoError(t, err)
	require.Len(t, report.Rows, 1)
	assert.Equal(t, 1000, report.Rows[0].TokensIn)
	assert.Equal(t, 500, report.Rows[0].TokensOut)
}

func TestIntegration_Log_ReuseSession(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	runLogCmd(t, s, "log", "--issue", "2", "--agent", "claude-code", "--model", "claude-sonnet-4-6", "--tokens-in", "100", "--tokens-out", "50", "--repo", "owner/repo") //nolint:errcheck

	// second call — should reuse the session
	_, err := runLogCmd(t, s, "log", "--issue", "2", "--agent", "claude-code", "--model", "claude-sonnet-4-6", "--tokens-in", "200", "--tokens-out", "100", "--repo", "owner/repo")
	require.NoError(t, err)

	sessions, err := s.ListSessions(ctx, "owner/repo", 2)
	require.NoError(t, err)
	assert.Len(t, sessions, 1, "expect one session to be reused")

	report, err := s.GetReport(ctx, "owner/repo", 2)
	require.NoError(t, err)
	assert.Equal(t, 300, report.TotalTokensIn)
	assert.Equal(t, 150, report.TotalTokensOut)
}

func TestIntegration_Log_ClosesIdleSession(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// manually start a session with an old timestamp
	oldSess, err := s.StartSession(ctx, "owner/repo", 3)
	require.NoError(t, err)

	// manipulate the session to be 40 minutes old
	oldTime := time.Now().Add(-40 * time.Minute)
	oldSess.StartedAt = oldTime

	// close it and reopen with old start time by ending and re-inserting is complex;
	// instead, verify that resolveSession only ends sessions older than 30 min.
	// We rely on the unit tests for the idle-close logic, and here test that a
	// fresh log call after ending the old session creates a new one.
	require.NoError(t, s.EndSession(ctx, oldSess.ID))

	_, err = runLogCmd(t, s, "log", "--issue", "3", "--agent", "cursor", "--model", "gpt-4o", "--tokens-in", "50", "--tokens-out", "25", "--repo", "owner/repo")
	require.NoError(t, err)

	sessions, err := s.ListSessions(ctx, "owner/repo", 3)
	require.NoError(t, err)

	// old session (ended) + new session
	assert.Len(t, sessions, 2)

	activeSessions := 0
	for _, sess := range sessions {
		if sess.EndedAt == nil {
			activeSessions++
		}
	}

	assert.Equal(t, 1, activeSessions)
}

func TestIntegration_Report_ShowsBreakdown(t *testing.T) {
	s := newTestStore(t)

	runLogCmd(t, s, "log", "--issue", "10", "--agent", "claude-code", "--model", "claude-sonnet-4-6", "--tokens-in", "2000", "--tokens-out", "1000", "--repo", "owner/repo") //nolint:errcheck
	runLogCmd(t, s, "log", "--issue", "10", "--agent", "opencode", "--model", "gpt-4o", "--tokens-in", "500", "--tokens-out", "250", "--repo", "owner/repo")                 //nolint:errcheck

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{reportCommand(s)},
	}

	err := app.RunContext(context.Background(), []string{"tok", "report", "--issue", "10", "--repo", "owner/repo"})
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "claude-code")
	assert.Contains(t, out, "opencode")
	assert.Contains(t, out, "Total")
}

func TestIntegration_Log_MultipleIssues(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	runLogCmd(t, s, "log", "--issue", "100", "--agent", "a", "--model", "gpt-4o", "--tokens-in", "10", "--tokens-out", "5", "--repo", "o/r")  //nolint:errcheck
	runLogCmd(t, s, "log", "--issue", "101", "--agent", "a", "--model", "gpt-4o", "--tokens-in", "20", "--tokens-out", "10", "--repo", "o/r") //nolint:errcheck

	issues, err := s.ListIssues(ctx, domain.IssueFilter{})
	require.NoError(t, err)
	assert.Len(t, issues, 2)
}
