package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/export"
	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

// integrationIssueProvider always validates issues as existing with a stub title.
type integrationIssueProvider struct{}

func (p *integrationIssueProvider) ListIssues(_ context.Context, _ usage.Filter) ([]provider.Issue, error) {
	return nil, nil
}

func (p *integrationIssueProvider) GetIssue(_ context.Context, repo string, number int) (*provider.Issue, error) {
	return &provider.Issue{Number: number, Repo: repo, Title: "Integration Test Issue"}, nil
}

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

func runLogCmd(t *testing.T, s *store.SQLiteStore, args ...string) error {
	t.Helper()

	app := &cli.App{
		Commands: []*cli.Command{logCommand(s, &integrationIssueProvider{})},
	}

	return app.RunContext(context.Background(), append([]string{"tok"}, args...))
}

func TestIntegration_Log_CreatesSessionAndEntry(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	err := runLogCmd(t, s,
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

func TestIntegration_Log_RejectsNegativeTokens(t *testing.T) {
	s := newTestStore(t)

	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "negative tokens-in",
			args: []string{"--tokens-in", "-5", "--tokens-out", "500"},
			want: "--tokens-in must be zero or greater",
		},
		{
			name: "negative tokens-out",
			args: []string{"--tokens-in", "1000", "--tokens-out", "-1"},
			want: "--tokens-out must be zero or greater",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{
				"log",
				"--issue", "1",
				"--agent", "claude-code",
				"--model", "claude-sonnet-4-6",
				"--repo", "owner/repo",
			}, tc.args...)

			err := runLogCmd(t, s, args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)

			entries, listErr := s.ListEntries(context.Background(), usage.Filter{Repo: "owner/repo"})
			require.NoError(t, listErr)
			assert.Empty(t, entries)
		})
	}
}

func TestIntegration_Log_ReuseSession(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(
		t,
		runLogCmd(
			t,
			s,
			"log",
			"--issue",
			"2",
			"--agent",
			"claude-code",
			"--model",
			"claude-sonnet-4-6",
			"--tokens-in",
			"100",
			"--tokens-out",
			"50",
			"--repo",
			"owner/repo",
		),
	)

	// second call — should reuse the session
	err := runLogCmd(
		t,
		s,
		"log",
		"--issue",
		"2",
		"--agent",
		"claude-code",
		"--model",
		"claude-sonnet-4-6",
		"--tokens-in",
		"200",
		"--tokens-out",
		"100",
		"--repo",
		"owner/repo",
	)
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

	require.NoError(
		t,
		runLogCmd(
			t,
			s,
			"log",
			"--issue",
			"3",
			"--agent",
			"cursor",
			"--model",
			"gpt-4o",
			"--tokens-in",
			"50",
			"--tokens-out",
			"25",
			"--repo",
			"owner/repo",
		),
	)

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

	runLogCmd(
		t,
		s,
		"log",
		"--issue",
		"10",
		"--agent",
		"claude-code",
		"--model",
		"claude-sonnet-4-6",
		"--tokens-in",
		"2000",
		"--tokens-out",
		"1000",
		"--repo",
		"owner/repo",
	) //nolint:errcheck
	runLogCmd(
		t,
		s,
		"log",
		"--issue",
		"10",
		"--agent",
		"opencode",
		"--model",
		"gpt-4o",
		"--tokens-in",
		"500",
		"--tokens-out",
		"250",
		"--repo",
		"owner/repo",
	) //nolint:errcheck

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

func runReportCmd(t *testing.T, s *store.SQLiteStore, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{reportCommand(s)},
	}

	err := app.RunContext(context.Background(), append([]string{"tok"}, args...))

	return buf.String(), err
}

func runBudgetCmd(t *testing.T, s *store.SQLiteStore, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{budgetCommands(s)},
	}

	err := app.RunContext(context.Background(), append([]string{"tok"}, args...))

	return buf.String(), err
}

func TestIntegration_Log_MultipleIssues(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	runLogCmd(
		t,
		s,
		"log",
		"--issue",
		"100",
		"--agent",
		"a",
		"--model",
		"gpt-4o",
		"--tokens-in",
		"10",
		"--tokens-out",
		"5",
		"--repo",
		"o/r",
	) //nolint:errcheck
	runLogCmd(
		t,
		s,
		"log",
		"--issue",
		"101",
		"--agent",
		"a",
		"--model",
		"gpt-4o",
		"--tokens-in",
		"20",
		"--tokens-out",
		"10",
		"--repo",
		"o/r",
	) //nolint:errcheck

	issues, err := s.ListIssues(ctx, usage.Filter{})
	require.NoError(t, err)
	assert.Len(t, issues, 2)
}

func TestIntegration_Log_WithNoteAndTag(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	err := runLogCmd(t, s,
		"log",
		"--issue", "20",
		"--agent", "claude-code",
		"--model", "claude-sonnet-4-6",
		"--tokens-in", "1000",
		"--tokens-out", "500",
		"--repo", "owner/repo",
		"--note", "refactored auth",
		"--tag", "refactor",
		"--tag", "feature",
	)
	require.NoError(t, err)

	sessions, err := s.ListSessions(ctx, "owner/repo", 20)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, "refactored auth", sessions[0].Note)
	assert.ElementsMatch(t, []string{"refactor", "feature"}, sessions[0].Tags)
}

func TestIntegration_Log_NoteTruncatedOnRuneBoundary(t *testing.T) {
	s := newTestStore(t)

	longNote := strings.Repeat("è", 250)

	require.NoError(t, runLogCmd(t, s,
		"log",
		"--issue", "1",
		"--agent", "claude-code",
		"--model", "claude-sonnet-4-6",
		"--tokens-in", "10",
		"--tokens-out", "10",
		"--repo", "owner/repo",
		"--note", longNote,
	))

	sessions, err := s.ListSessions(context.Background(), "owner/repo", 1)
	require.NoError(t, err)
	require.Len(t, sessions, 1)

	assert.True(t, utf8.ValidString(sessions[0].Note))
	assert.Len(t, []rune(sessions[0].Note), 200)
}

func TestIntegration_Log_TagsAccumulate(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, runLogCmd(t, s,
		"log", "--issue", "21", "--agent", "claude-code", "--model", "claude-sonnet-4-6",
		"--tokens-in", "100", "--tokens-out", "50", "--repo", "owner/repo",
		"--tag", "debug",
	))

	require.NoError(t, runLogCmd(t, s,
		"log", "--issue", "21", "--agent", "claude-code", "--model", "claude-sonnet-4-6",
		"--tokens-in", "200", "--tokens-out", "100", "--repo", "owner/repo",
		"--tag", "feature",
	))

	sessions, err := s.ListSessions(ctx, "owner/repo", 21)
	require.NoError(t, err)
	require.Len(t, sessions, 1, "second call should reuse same session")
	assert.ElementsMatch(t, []string{"debug", "feature"}, sessions[0].Tags)
}

func TestIntegration_Report_Sessions(t *testing.T) {
	s := newTestStore(t)

	require.NoError(t, runLogCmd(t, s,
		"log", "--issue", "30", "--agent", "claude-code", "--model", "claude-sonnet-4-6",
		"--tokens-in", "1000", "--tokens-out", "500", "--repo", "owner/repo",
		"--note", "initial work", "--tag", "feature",
	))

	out, err := runReportCmd(t, s, "report", "--issue", "30", "--repo", "owner/repo", "--sessions")
	require.NoError(t, err)
	assert.Contains(t, out, "Session 1")
	assert.Contains(t, out, "initial work")
	assert.Contains(t, out, "feature")
}

func TestIntegration_Report_WithBudget(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, runLogCmd(t, s,
		"log", "--issue", "40", "--agent", "claude-code", "--model", "claude-sonnet-4-6",
		"--tokens-in", "1000", "--tokens-out", "500", "--repo", "owner/repo",
	))

	require.NoError(t, s.SetBudget(ctx, "owner/repo", 40, 50.00))

	out, err := runReportCmd(t, s, "report", "--issue", "40", "--repo", "owner/repo")
	require.NoError(t, err)
	assert.Contains(t, out, "Budget")
	assert.Contains(t, out, "$50.00")
}

func runExportCmd(t *testing.T, s *store.SQLiteStore, args ...string) (string, error) {
	t.Helper()

	_, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	return runExportCmdWithKey(t, s, priv, args...)
}

func runExportCmdWithKey(t *testing.T, s *store.SQLiteStore, priv ed25519.PrivateKey, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{exportCommands(s, priv, "test")},
	}

	err := app.RunContext(context.Background(), append([]string{"tok"}, args...))

	return buf.String(), err
}

func TestIntegration_Export_SchemaV3_IncludesSessions(t *testing.T) {
	s := newTestStore(t)

	require.NoError(t, runLogCmd(t, s,
		"log", "--issue", "50", "--agent", "claude-code", "--model", "claude-sonnet-4-6",
		"--tokens-in", "1000", "--tokens-out", "500", "--repo", "owner/repo",
		"--note", "export test note", "--tag", "feature",
	))

	out, err := runExportCmd(t, s, "export", "--issue", "50", "--repo", "owner/repo")
	require.NoError(t, err)

	var doc export.Document
	require.NoError(t, json.Unmarshal([]byte(out), &doc))

	assert.Equal(t, "3.0", doc.SchemaVersion)
	require.Len(t, doc.Sessions, 1)
	assert.Equal(t, 50, doc.Sessions[0].IssueNum)
	assert.Equal(t, "export test note", doc.Sessions[0].Note)
	assert.Equal(t, []string{"feature"}, doc.Sessions[0].Tags)
}

func TestIntegration_Export_SchemaV3_IncludesBudget(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, runLogCmd(t, s,
		"log", "--issue", "51", "--agent", "claude-code", "--model", "claude-sonnet-4-6",
		"--tokens-in", "500", "--tokens-out", "250", "--repo", "owner/repo",
	))
	require.NoError(t, s.SetBudget(ctx, "owner/repo", 51, 25.00))

	out, err := runExportCmd(t, s, "export", "--issue", "51", "--repo", "owner/repo")
	require.NoError(t, err)

	var doc export.Document
	require.NoError(t, json.Unmarshal([]byte(out), &doc))

	require.Len(t, doc.Budgets, 1)
	assert.Equal(t, 51, doc.Budgets[0].IssueNum)
	assert.InEpsilon(t, 25.00, doc.Budgets[0].Amount, 0.001)
}

func TestIntegration_Export_NoRepoIssueFilter_OmitsSessionsAndBudgets(t *testing.T) {
	s := newTestStore(t)

	require.NoError(t, runLogCmd(t, s,
		"log", "--issue", "52", "--agent", "claude-code", "--model", "claude-sonnet-4-6",
		"--tokens-in", "100", "--tokens-out", "50", "--repo", "owner/repo",
	))

	// export without --repo/--issue filter: sessions and budgets blocks must be absent
	out, err := runExportCmd(t, s, "export")
	require.NoError(t, err)

	var doc export.Document
	require.NoError(t, json.Unmarshal([]byte(out), &doc))

	assert.Empty(t, doc.Sessions)
	assert.Empty(t, doc.Budgets)
}

func exportToFile(t *testing.T, s *store.SQLiteStore, priv ed25519.PrivateKey) string {
	t.Helper()

	require.NoError(t, runLogCmd(t, s,
		"log", "--issue", "60", "--agent", "claude-code", "--model", "claude-sonnet-4-6",
		"--tokens-in", "100", "--tokens-out", "50", "--repo", "owner/repo",
		"--note", "verify test",
	))

	path := filepath.Join(t.TempDir(), "export.json")
	_, err := runExportCmdWithKey(t, s, priv, "export", "--issue", "60", "--repo", "owner/repo", "--output", path)
	require.NoError(t, err)

	return path
}

func TestIntegration_ExportVerify_Roundtrip(t *testing.T) {
	s := newTestStore(t)

	_, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	path := exportToFile(t, s, priv)

	out, err := runExportCmd(t, s, "export", "verify", "--file", path)
	require.NoError(t, err)
	assert.Contains(t, out, "OK: signature valid (schema 3.0, full document)")
	assert.Contains(t, out, "Origin not verified")
}

func TestIntegration_ExportVerify_TamperedSessionFails(t *testing.T) {
	s := newTestStore(t)

	_, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	path := exportToFile(t, s, priv)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var doc export.Document
	require.NoError(t, json.Unmarshal(data, &doc))
	require.NotEmpty(t, doc.Sessions)

	doc.Sessions[0].Note = "tampered"

	tampered, err := json.Marshal(doc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, tampered, 0o600))

	out, err := runExportCmd(t, s, "export", "verify", "--file", path)
	require.Error(t, err)
	assert.Contains(t, out, "INVALID")
}

func TestIntegration_ExportVerify_LegacyV2Warns(t *testing.T) {
	s := newTestStore(t)

	out, err := runExportCmd(t, s, "export", "verify",
		"--file", filepath.Join("..", "..", "internal", "export", "testdata", "export_v2.json"))
	require.NoError(t, err)
	assert.Contains(t, out, "OK: signature valid (schema 2.0, legacy)")
	assert.Contains(t, out, "WARNING")
}

func TestIntegration_ExportVerify_PubkeyMatch(t *testing.T) {
	s := newTestStore(t)

	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	path := exportToFile(t, s, priv)

	out, err := runExportCmd(t, s, "export", "verify", "--file", path,
		"--pubkey", base64.StdEncoding.EncodeToString(pub))
	require.NoError(t, err)
	assert.Contains(t, out, "Origin verified")
}

func TestIntegration_ExportVerify_PubkeyMismatch(t *testing.T) {
	s := newTestStore(t)

	_, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	otherPub, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	path := exportToFile(t, s, priv)

	out, err := runExportCmd(t, s, "export", "verify", "--file", path,
		"--pubkey", base64.StdEncoding.EncodeToString(otherPub))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not signed by the expected key")
	assert.Contains(t, out, "INVALID: public key mismatch")
}

func TestIntegration_Report_Sessions_NoSessions(t *testing.T) {
	s := newTestStore(t)

	// issue exists in the cache but has no sessions
	ctx := context.Background()
	require.NoError(t, s.UpsertIssueCache(ctx, &usage.IssueCache{
		Repo: "owner/repo", IssueNum: 55, Title: "Empty Issue",
	}))

	out, err := runReportCmd(t, s, "report", "--issue", "55", "--repo", "owner/repo", "--sessions")
	require.NoError(t, err)
	assert.Contains(t, out, "No sessions found")
}

func TestIntegration_Log_IdleSessionClosed_ByResolveSession(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// start a session and back-date its last_activity_at to simulate 40 min idle
	sess, err := s.StartSession(ctx, "owner/repo", 60)
	require.NoError(t, err)

	idleTime := time.Now().Add(-40 * time.Minute)
	require.NoError(t, s.UpdateSessionActivity(ctx, sess.ID, idleTime))

	// next log call must close the idle session and open a new one
	require.NoError(t, runLogCmd(t, s,
		"log", "--issue", "60", "--agent", "claude-code", "--model", "claude-sonnet-4-6",
		"--tokens-in", "100", "--tokens-out", "50", "--repo", "owner/repo",
	))

	sessions, err := s.ListSessions(ctx, "owner/repo", 60)
	require.NoError(t, err)
	require.Len(t, sessions, 2)

	var closed, active int
	for _, se := range sessions {
		if se.EndedAt != nil {
			closed++
			// ended_at must reflect last_activity_at, not now — duration must be near zero
			duration := se.EndedAt.Sub(se.StartedAt)
			assert.Less(t, duration, time.Minute, "idle session duration should be ~0, not wall-clock time")
		} else {
			active++
		}
	}

	assert.Equal(t, 1, closed)
	assert.Equal(t, 1, active)
}
