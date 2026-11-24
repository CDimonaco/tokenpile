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

func TestSQLiteStore_UpsertCapturedSession(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	t0 := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	issue := 7

	require.NoError(t, s.UpsertCapturedSession(ctx, "s1", "o/r", nil, t0.Add(time.Minute), t0.Add(2*time.Minute)))
	require.NoError(t, s.UpsertCapturedSession(ctx, "s1", "o/r", &issue, t0, t0.Add(5*time.Minute)))
	// A later unattributed replay must not erase the known issue.
	require.NoError(t, s.UpsertCapturedSession(ctx, "s1", "o/r", nil, t0.Add(time.Minute), t0.Add(time.Minute)))

	sessions, err := s.ListAllSessions(ctx)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, t0, sessions[0].StartedAt)
	require.NotNil(t, sessions[0].EndedAt)
	assert.Equal(t, t0.Add(5*time.Minute), *sessions[0].EndedAt)
	require.NotNil(t, sessions[0].IssueNum)
	assert.Equal(t, 7, *sessions[0].IssueNum)
}

func TestSQLiteStore_BackfillsOrphanSessions(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	ctx := context.Background()
	pricer := fixedPricer{pricePerMIn: 3.0, pricePerMOut: 15.0}
	t0 := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	issue := 3

	s, err := store.NewSQLiteStore(dbPath, pricer)
	require.NoError(t, err)

	for i, at := range []time.Time{t0, t0.Add(time.Hour)} {
		require.NoError(t, s.LogUsage(ctx, usage.Entry{
			ID: string(rune('a' + i)), Repo: "o/r", IssueNum: &issue, Agent: "claude-code", Model: "m",
			SessionID: "orphan", Source: usage.SourceMeasured, At: at,
			Usage: usage.Usage{Output: 1},
		}))
	}

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		ID: "c", Repo: "o/r", Agent: "claude-code", Model: "m",
		SessionID: "existing", Source: usage.SourceMeasured, At: t0,
		Usage: usage.Usage{Output: 1},
	}))
	require.NoError(t, s.UpsertCapturedSession(ctx, "existing", "o/r", nil, t0, t0))
	require.NoError(t, s.Close())

	for range 2 {
		s, err = store.NewSQLiteStore(dbPath, pricer)
		require.NoError(t, err)

		sessions, listErr := s.ListAllSessions(ctx)
		require.NoError(t, listErr)
		require.Len(t, sessions, 2)

		for _, sess := range sessions {
			if sess.ID != "orphan" {
				continue
			}

			assert.Equal(t, t0, sess.StartedAt)
			require.NotNil(t, sess.EndedAt)
			assert.Equal(t, t0.Add(time.Hour), *sess.EndedAt)
			require.NotNil(t, sess.IssueNum)
			assert.Equal(t, 3, *sess.IssueNum)
		}

		require.NoError(t, s.Close())
	}
}
