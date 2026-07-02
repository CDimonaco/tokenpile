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

	"github.com/cdimonaco/tokenpile/internal/domain"
	"github.com/cdimonaco/tokenpile/internal/mocks"
)

func runLogApp(t *testing.T, storeMock *mocks.Store, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{logCommand(storeMock)},
	}

	err := app.RunContext(context.Background(), append([]string{"tok"}, args...))

	return buf.String(), err
}

func TestLog_MissingRequiredFlags(t *testing.T) {
	s := &mocks.Store{}

	_, err := runLogApp(t, s, "log", "--agent", "claude-code", "--model", "claude-sonnet-4-6")

	assert.Error(t, err)
}

func TestLog_NoActiveSession_StartsNew(t *testing.T) {
	s := &mocks.Store{}

	sess := &domain.Session{
		ID:        "sess-1",
		Repo:      "owner/repo",
		IssueNum:  42,
		StartedAt: time.Now(),
	}

	s.On("ListSessions", mock.Anything, "owner/repo", 42).Return([]domain.Session{}, nil)
	s.On("StartSession", mock.Anything, "owner/repo", 42).Return(sess, nil)
	s.On("LogUsage", mock.Anything, mock.MatchedBy(func(e domain.UsageEntry) bool {
		return e.Repo == "owner/repo" && e.IssueNum == 42 && e.SessionID == "sess-1"
	})).Return(nil)

	out, err := runLogApp(t, s,
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
}

func TestLog_ReuseActiveSession(t *testing.T) {
	s := &mocks.Store{}

	recentTime := time.Now().Add(-5 * time.Minute)
	activeSess := domain.Session{
		ID:        "sess-active",
		Repo:      "owner/repo",
		IssueNum:  7,
		StartedAt: recentTime,
	}

	s.On("ListSessions", mock.Anything, "owner/repo", 7).Return([]domain.Session{activeSess}, nil)
	s.On("LogUsage", mock.Anything, mock.MatchedBy(func(e domain.UsageEntry) bool {
		return e.SessionID == "sess-active"
	})).Return(nil)

	_, err := runLogApp(t, s,
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
}

func TestLog_ClosesIdleSession_StartsNew(t *testing.T) {
	s := &mocks.Store{}

	idleTime := time.Now().Add(-45 * time.Minute)
	idleSess := domain.Session{
		ID:        "sess-idle",
		Repo:      "owner/repo",
		IssueNum:  99,
		StartedAt: idleTime,
	}

	newSess := &domain.Session{
		ID:        "sess-new",
		Repo:      "owner/repo",
		IssueNum:  99,
		StartedAt: time.Now(),
	}

	s.On("ListSessions", mock.Anything, "owner/repo", 99).Return([]domain.Session{idleSess}, nil)
	s.On("EndSession", mock.Anything, "sess-idle").Return(nil)
	s.On("StartSession", mock.Anything, "owner/repo", 99).Return(newSess, nil)
	s.On("LogUsage", mock.Anything, mock.MatchedBy(func(e domain.UsageEntry) bool {
		return e.SessionID == "sess-new"
	})).Return(nil)

	_, err := runLogApp(t, s,
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
}
