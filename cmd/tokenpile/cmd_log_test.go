package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/mocks"
	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

func runLogApp(t *testing.T, storeMock *mocks.Store, ip *mocks.IssueProvider, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{logCommand(storeMock, ip)},
	}

	err := app.RunContext(context.Background(), append([]string{"tok"}, args...))

	return buf.String(), err
}

func stubIssue(ip *mocks.IssueProvider, repo string, num int) {
	ip.On("GetIssue", mock.Anything, repo, num).Return(&provider.Issue{
		Number: num,
		Repo:   repo,
		Title:  "Stub Issue",
	}, nil)
}

func TestLog_MissingRequiredFlags(t *testing.T) {
	s := &mocks.Store{}
	ip := &mocks.IssueProvider{}

	_, err := runLogApp(t, s, ip, "log", "--agent", "claude-code", "--model", "claude-sonnet-4-6")

	assert.Error(t, err)
}

func TestLog_NoActiveSession_StartsNew(t *testing.T) {
	s := &mocks.Store{}
	ip := &mocks.IssueProvider{}

	stubIssue(ip, "owner/repo", 42)
	s.On("UpsertIssueCache", mock.Anything, mock.Anything).Return(nil)

	sess := &usage.Session{
		ID:        "sess-1",
		Repo:      "owner/repo",
		IssueNum:  42,
		StartedAt: time.Now(),
	}

	s.On("ListSessions", mock.Anything, "owner/repo", 42).Return([]usage.Session{}, nil)
	s.On("StartSession", mock.Anything, "owner/repo", 42).Return(sess, nil)
	s.On("UpdateSessionActivity", mock.Anything, "sess-1", mock.AnythingOfType("time.Time")).Return(nil)
	s.On("LogUsage", mock.Anything, mock.MatchedBy(func(e usage.Entry) bool {
		return e.Repo == "owner/repo" && e.IssueNum == 42 && e.SessionID == "sess-1"
	})).Return(nil)

	out, err := runLogApp(t, s, ip,
		"log",
		"--issue", "42",
		"--agent", "claude-code",
		"--model", "claude-sonnet-4-6",
		"--tokens-in", "1000",
		"--tokens-out", "500",
		"--repo", "owner/repo",
	)

	require.NoError(t, err)
	assert.Contains(t, out, "owner/repo")
	assert.Contains(t, out, "42")
	s.AssertExpectations(t)
	ip.AssertExpectations(t)
}

func TestLog_ReuseActiveSession(t *testing.T) {
	s := &mocks.Store{}
	ip := &mocks.IssueProvider{}

	stubIssue(ip, "owner/repo", 7)
	s.On("UpsertIssueCache", mock.Anything, mock.Anything).Return(nil)

	recentTime := time.Now().Add(-5 * time.Minute)
	activeSess := usage.Session{
		ID:             "sess-active",
		Repo:           "owner/repo",
		IssueNum:       7,
		StartedAt:      recentTime,
		LastActivityAt: recentTime,
	}

	s.On("ListSessions", mock.Anything, "owner/repo", 7).Return([]usage.Session{activeSess}, nil)
	s.On("UpdateSessionActivity", mock.Anything, "sess-active", mock.AnythingOfType("time.Time")).Return(nil)
	s.On("LogUsage", mock.Anything, mock.MatchedBy(func(e usage.Entry) bool {
		return e.SessionID == "sess-active"
	})).Return(nil)

	_, err := runLogApp(t, s, ip,
		"log",
		"--issue", "7",
		"--agent", "opencode",
		"--model", "gpt-4o",
		"--tokens-in", "200",
		"--tokens-out", "100",
		"--repo", "owner/repo",
	)

	require.NoError(t, err)
	s.AssertExpectations(t)
	ip.AssertExpectations(t)
}

func TestLog_ClosesIdleSession_StartsNew(t *testing.T) {
	s := &mocks.Store{}
	ip := &mocks.IssueProvider{}

	stubIssue(ip, "owner/repo", 99)
	s.On("UpsertIssueCache", mock.Anything, mock.Anything).Return(nil)

	idleTime := time.Now().Add(-45 * time.Minute)
	idleSess := usage.Session{
		ID:             "sess-idle",
		Repo:           "owner/repo",
		IssueNum:       99,
		StartedAt:      idleTime,
		LastActivityAt: idleTime,
	}

	newSess := &usage.Session{
		ID:        "sess-new",
		Repo:      "owner/repo",
		IssueNum:  99,
		StartedAt: time.Now(),
	}

	s.On("ListSessions", mock.Anything, "owner/repo", 99).Return([]usage.Session{idleSess}, nil)
	s.On("EndSessionAt", mock.Anything, "sess-idle", mock.AnythingOfType("time.Time")).Return(nil)
	s.On("StartSession", mock.Anything, "owner/repo", 99).Return(newSess, nil)
	s.On("UpdateSessionActivity", mock.Anything, "sess-new", mock.AnythingOfType("time.Time")).Return(nil)
	s.On("LogUsage", mock.Anything, mock.MatchedBy(func(e usage.Entry) bool {
		return e.SessionID == "sess-new"
	})).Return(nil)

	_, err := runLogApp(t, s, ip,
		"log",
		"--issue", "99",
		"--agent", "cursor",
		"--model", "claude-opus-4-7",
		"--tokens-in", "300",
		"--tokens-out", "150",
		"--repo", "owner/repo",
	)

	require.NoError(t, err)
	s.AssertExpectations(t)
	ip.AssertExpectations(t)
}
