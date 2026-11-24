package store_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/cdimonaco/tokenpile/internal/store"
)

func TestSQLiteStore_TouchCapturedSession_WidensRange(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	t0 := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	require.NoError(t, s.TouchCapturedSession(ctx, "s1", "o/r", nil, t0))
	require.NoError(t, s.TouchCapturedSession(ctx, "s1", "o/r", nil, t0.Add(time.Hour)))
	require.NoError(t, s.TouchCapturedSession(ctx, "s1", "o/r", nil, t0.Add(-time.Hour)))
	require.NoError(t, s.TouchCapturedSession(ctx, "s1", "o/r", nil, t0))

	sessions, err := s.ListSessions(ctx, "o/r", nil)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, "s1", sessions[0].ID)
	assert.True(t, sessions[0].StartedAt.Equal(t0.Add(-time.Hour)))
	require.NotNil(t, sessions[0].EndedAt)
	assert.True(t, sessions[0].EndedAt.Equal(t0.Add(time.Hour)))
	assert.True(t, sessions[0].LastActivityAt.Equal(t0.Add(time.Hour)))
}

func TestSQLiteStore_TouchCapturedSession_NilIssueKeepsExisting(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	at := time.Now().UTC()

	require.NoError(t, s.TouchCapturedSession(ctx, "s1", "o/r", issuePtr(42), at))
	require.NoError(t, s.TouchCapturedSession(ctx, "s1", "o/r", nil, at.Add(time.Minute)))

	sessions, err := s.ListSessions(ctx, "o/r", issuePtr(42))
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
}

func TestSQLiteStore_BackfillsCapturedSessions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")

	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)

	_, err = db.Exec(`
CREATE TABLE usage_entries (
    id TEXT PRIMARY KEY, repo TEXT NOT NULL, issue_num INTEGER,
    agent TEXT NOT NULL, model TEXT NOT NULL,
    input_fresh INTEGER NOT NULL, cache_write INTEGER NOT NULL,
    cache_read INTEGER NOT NULL, output INTEGER NOT NULL, reasoning INTEGER NOT NULL,
    source TEXT NOT NULL, session_id TEXT, at TEXT NOT NULL);
INSERT INTO usage_entries VALUES
    ('e1','o/r',42,'claude-code','m',10,0,0,5,0,'measured','s1','2026-07-01T00:00:00Z'),
    ('e2','o/r',42,'claude-code','m',10,0,0,5,0,'measured','s1','2026-07-03T00:00:00Z'),
    ('e3','o/r',NULL,'claude-code','m',10,0,0,5,0,'measured',NULL,'2026-07-03T00:00:00Z');`)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	loader, err := pricing.NewLoader("")
	require.NoError(t, err)

	for range 2 { // reopening must not duplicate
		s, openErr := store.NewSQLiteStore(path, loader)
		require.NoError(t, openErr)

		sessions, listErr := s.ListSessions(context.Background(), "o/r", issuePtr(42))
		require.NoError(t, listErr)
		require.Len(t, sessions, 1)
		assert.Equal(t, "s1", sessions[0].ID)
		assert.Equal(t, 1, sessions[0].StartedAt.Day())
		require.NotNil(t, sessions[0].EndedAt)
		assert.Equal(t, 3, sessions[0].EndedAt.Day())

		all, allErr := s.ListAllSessions(context.Background())
		require.NoError(t, allErr)
		assert.Len(t, all, 1)

		require.NoError(t, s.Close())
	}
}
