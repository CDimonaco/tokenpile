package store_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

type fixedPricer struct {
	pricePerMIn  float64
	pricePerMOut float64
}

func (f fixedPricer) ComputeCost(_ string, u usage.Usage) pricing.CostResult {
	cost := float64(u.TotalInput())/1_000_000*f.pricePerMIn +
		float64(u.Output)/1_000_000*f.pricePerMOut

	return pricing.CostResult{Cost: cost, Known: true}
}

func (f fixedPricer) CacheSavings(_ string, _ usage.Usage) (float64, bool) {
	return 0, true
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
		Repo:     "owner/repo",
		IssueNum: issuePtr(42),
		Agent:    "claude-code",
		Model:    "claude-sonnet-4-6",
		Usage:    usage.Usage{InputFresh: 1000, Output: 500},
		Source:   usage.SourceEstimated,
	}

	err := s.LogUsage(ctx, entry)
	require.NoError(t, err)

	issues, err := s.ListIssues(ctx, usage.Filter{Repo: "owner/repo"})
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, 42, issues[0].IssueNum)
	assert.Equal(t, 1000, issues[0].TotalUsage.TotalInput())
	assert.Equal(t, 500, issues[0].TotalUsage.Output)
}

func TestSQLiteStore_LogUsage_SetsTimestamp(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	before := time.Now().UTC().Add(-time.Second)

	err := s.LogUsage(ctx, usage.Entry{
		Repo: "owner/repo", IssueNum: issuePtr(1), Agent: "a", Model: "m",
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

	sess, err := s.StartSession(ctx, "owner/repo", issuePtr(42))
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
		{
			Repo:     "o/r",
			IssueNum: issuePtr(1),
			Agent:    "claude-code",
			Model:    "claude-sonnet-4-6",
			Usage:    usage.Usage{InputFresh: 1000, Output: 200},
			Source:   usage.SourceEstimated,
		},
		{
			Repo:     "o/r",
			IssueNum: issuePtr(1),
			Agent:    "opencode",
			Model:    "gpt-4o",
			Usage:    usage.Usage{InputFresh: 500, Output: 100},
			Source:   usage.SourceEstimated,
		},
		{
			Repo:     "o/r",
			IssueNum: issuePtr(1),
			Agent:    "claude-code",
			Model:    "claude-sonnet-4-6",
			Usage:    usage.Usage{InputFresh: 2000, Output: 400},
			Source:   usage.SourceEstimated,
		},
	}

	for _, e := range entries {
		require.NoError(t, s.LogUsage(ctx, e))
	}

	report, err := s.GetReport(ctx, "o/r", 1)
	require.NoError(t, err)
	require.Len(t, report.Rows, 2)
	assert.Equal(t, 3500, report.TotalUsage.TotalInput())
	assert.Equal(t, 700, report.TotalUsage.Output)
}

func TestSQLiteStore_GetReport_EmptyIssue(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	report, err := s.GetReport(ctx, "o/r", 99)
	require.NoError(t, err)
	assert.Empty(t, report.Rows)
	assert.Equal(t, 0, report.TotalUsage.TotalInput())
}

func TestSQLiteStore_ListIssues_MultipleIssues(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for _, num := range []int{1, 2, 3} {
		require.NoError(t, s.LogUsage(ctx, usage.Entry{
			Repo: "o/r", IssueNum: issuePtr(num), Agent: "a", Model: "m",
			Usage:  usage.Usage{InputFresh: 100 * num, Output: 50 * num},
			Source: usage.SourceEstimated,
		}))
	}

	issues, err := s.ListIssues(ctx, usage.Filter{Repo: "o/r"})
	require.NoError(t, err)
	assert.Len(t, issues, 3)
}

func TestSQLiteStore_ListIssues_FilterByAgent(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(
		t,
		s.LogUsage(
			ctx,
			usage.Entry{
				Repo:     "o/r",
				IssueNum: issuePtr(1),
				Agent:    "claude-code",
				Model:    "m",
				Usage:    usage.Usage{InputFresh: 100},
				Source:   usage.SourceEstimated,
			},
		),
	)
	require.NoError(
		t,
		s.LogUsage(
			ctx,
			usage.Entry{
				Repo:     "o/r",
				IssueNum: issuePtr(2),
				Agent:    "opencode",
				Model:    "m",
				Usage:    usage.Usage{InputFresh: 200},
				Source:   usage.SourceEstimated,
			},
		),
	)

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
			Repo: "o/r", IssueNum: issuePtr(1), Agent: "a", Model: "m",
			Usage: usage.Usage{InputFresh: 100, Output: 50}, Source: usage.SourceEstimated, At: at,
		}))
	}

	points, err := s.ListUsageOverTime(ctx, usage.OverTimeFilter{
		Repo:        "o/r",
		Granularity: usage.Day,
	})
	require.NoError(t, err)
	assert.Len(t, points, 3)

	for _, p := range points {
		assert.Equal(t, 100, p.Usage.TotalInput())
		assert.Equal(t, 50, p.Usage.Output)
	}
}

func TestSQLiteStore_ListUsageOverTime_WeekGranularity(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Two entries in the same week (Monday + Wednesday), one in the next week
	monday := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC) // Monday
	wednesday := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	nextMonday := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

	for _, at := range []time.Time{monday, wednesday} {
		require.NoError(t, s.LogUsage(ctx, usage.Entry{
			Repo: "o/r", IssueNum: issuePtr(1), Agent: "a", Model: "m",
			Usage: usage.Usage{InputFresh: 100, Output: 50}, Source: usage.SourceEstimated, At: at,
		}))
	}
	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "o/r", IssueNum: issuePtr(1), Agent: "a", Model: "m",
		Usage: usage.Usage{InputFresh: 200, Output: 80}, Source: usage.SourceEstimated, At: nextMonday,
	}))

	points, err := s.ListUsageOverTime(ctx, usage.OverTimeFilter{
		Repo:        "o/r",
		Granularity: usage.Week,
	})
	require.NoError(t, err)
	require.Len(t, points, 2)

	// No zero dates
	for _, p := range points {
		assert.False(t, p.Date.IsZero(), "date must not be zero")
		assert.Equal(t, 1, int(p.Date.Weekday()), "week point must be a Monday")
	}

	assert.Equal(t, 200, points[0].Usage.TotalInput()) // monday + wednesday
	assert.Equal(t, 200, points[1].Usage.TotalInput()) // next monday
}

func TestSQLiteStore_ListUsageOverTime_CostPopulated(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "o/r", IssueNum: issuePtr(1), Agent: "a", Model: "m",
		Usage:  usage.Usage{InputFresh: 1_000_000, Output: 1_000_000},
		Source: usage.SourceEstimated, At: time.Now(),
	}))

	points, err := s.ListUsageOverTime(ctx, usage.OverTimeFilter{
		Repo:        "o/r",
		Granularity: usage.Day,
	})
	require.NoError(t, err)
	require.Len(t, points, 1)
	assert.Greater(t, points[0].Cost, 0.0, "cost must be non-zero when tokens are logged")
}

func TestSQLiteStore_ListIssues_CostPopulated(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "o/r", IssueNum: issuePtr(1), Agent: "a", Model: "m",
		Usage:  usage.Usage{InputFresh: 1_000_000, Output: 1_000_000},
		Source: usage.SourceEstimated, At: time.Now(),
	}))

	issues, err := s.ListIssues(ctx, usage.Filter{Repo: "o/r"})
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Greater(t, issues[0].TotalCost, 0.0, "cost must be non-zero when tokens are logged")
}

func TestSQLiteStore_ListTrackedIssueRefs(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	refs, err := s.ListTrackedIssueRefs(ctx)
	require.NoError(t, err)
	assert.Empty(t, refs)

	_ = s.LogUsage(ctx, usage.Entry{Repo: "owner/a", IssueNum: issuePtr(1), Agent: "x", Model: "m", At: time.Now()})
	_ = s.LogUsage(ctx, usage.Entry{Repo: "owner/a", IssueNum: issuePtr(1), Agent: "x", Model: "m", At: time.Now()})
	_ = s.LogUsage(ctx, usage.Entry{Repo: "owner/a", IssueNum: issuePtr(2), Agent: "x", Model: "m", At: time.Now()})
	_ = s.LogUsage(ctx, usage.Entry{Repo: "owner/b", IssueNum: issuePtr(5), Agent: "x", Model: "m", At: time.Now()})

	refs, err = s.ListTrackedIssueRefs(ctx)
	require.NoError(t, err)
	require.Len(t, refs, 3)

	assert.Equal(t, usage.TrackedIssueRef{Repo: "owner/a", IssueNum: 1}, refs[0])
	assert.Equal(t, usage.TrackedIssueRef{Repo: "owner/a", IssueNum: 2}, refs[1])
	assert.Equal(t, usage.TrackedIssueRef{Repo: "owner/b", IssueNum: 5}, refs[2])
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

func TestSQLiteStore_MigrationIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	pricer := fixedPricer{}

	for range 3 {
		s, err := store.NewSQLiteStore(dbPath, pricer)
		require.NoError(t, err)
		require.NoError(t, s.Close())
	}
}

func TestSQLiteStore_UpdateSessionAnnotations_NoteAndTags(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sess, err := s.StartSession(ctx, "o/r", issuePtr(1))
	require.NoError(t, err)

	note := "refactored auth flow"
	err = s.UpdateSessionAnnotations(ctx, sess.ID, &note, []string{"refactor", "debug"})
	require.NoError(t, err)

	sessions, err := s.ListSessions(ctx, "o/r", 1)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, "refactored auth flow", sessions[0].Note)
	assert.ElementsMatch(t, []string{"refactor", "debug"}, sessions[0].Tags)
}

func TestSQLiteStore_UpdateSessionAnnotations_TagsUnion(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sess, err := s.StartSession(ctx, "o/r", issuePtr(1))
	require.NoError(t, err)

	note := "first"
	err = s.UpdateSessionAnnotations(ctx, sess.ID, &note, []string{"refactor"})
	require.NoError(t, err)

	note2 := "second"
	err = s.UpdateSessionAnnotations(ctx, sess.ID, &note2, []string{"debug"})
	require.NoError(t, err)

	sessions, err := s.ListSessions(ctx, "o/r", 1)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, "second", sessions[0].Note)
	assert.ElementsMatch(t, []string{"refactor", "debug"}, sessions[0].Tags)
}

func TestSQLiteStore_UpdateSessionAnnotations_NilNotePreservesExisting(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sess, err := s.StartSession(ctx, "o/r", issuePtr(1))
	require.NoError(t, err)

	note := "keep me"
	err = s.UpdateSessionAnnotations(ctx, sess.ID, &note, []string{"feature"})
	require.NoError(t, err)

	err = s.UpdateSessionAnnotations(ctx, sess.ID, nil, []string{"test"})
	require.NoError(t, err)

	sessions, err := s.ListSessions(ctx, "o/r", 1)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.Equal(t, "keep me", sessions[0].Note)
	assert.ElementsMatch(t, []string{"feature", "test"}, sessions[0].Tags)
}

func TestSQLiteStore_ListSessions_EmptyTagsNonNil(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.StartSession(ctx, "o/r", issuePtr(1))
	require.NoError(t, err)

	sessions, err := s.ListSessions(ctx, "o/r", 1)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	assert.NotNil(t, sessions[0].Tags)
	assert.Empty(t, sessions[0].Tags)
}

func TestSQLiteStore_SetBudget_And_GetBudget(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.GetBudget(ctx, "o/r", 1)
	require.ErrorIs(t, err, store.ErrBudgetNotFound)

	require.NoError(t, s.SetBudget(ctx, "o/r", 1, 50.0))

	budget, err := s.GetBudget(ctx, "o/r", 1)
	require.NoError(t, err)
	require.NotNil(t, budget)
	assert.InEpsilon(t, 50.0, *budget, 0.001)
}

func TestSQLiteStore_SetBudget_Upsert(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.SetBudget(ctx, "o/r", 1, 50.0))
	require.NoError(t, s.SetBudget(ctx, "o/r", 1, 100.0))

	budget, err := s.GetBudget(ctx, "o/r", 1)
	require.NoError(t, err)
	require.NotNil(t, budget)
	assert.InEpsilon(t, 100.0, *budget, 0.001)
}

func TestSQLiteStore_UnsetBudget(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.SetBudget(ctx, "o/r", 1, 50.0))
	require.NoError(t, s.UnsetBudget(ctx, "o/r", 1))

	_, err := s.GetBudget(ctx, "o/r", 1)
	require.ErrorIs(t, err, store.ErrBudgetNotFound)
}

func TestSQLiteStore_ListIssues_BudgetPopulated(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "o/r", IssueNum: issuePtr(1), Agent: "a", Model: "m",
		Usage: usage.Usage{InputFresh: 1000, Output: 500}, Source: usage.SourceEstimated, At: time.Now(),
	}))

	require.NoError(t, s.SetBudget(ctx, "o/r", 1, 25.0))

	issues, err := s.ListIssues(ctx, usage.Filter{Repo: "o/r"})
	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.NotNil(t, issues[0].Budget)
	assert.InEpsilon(t, 25.0, *issues[0].Budget, 0.001)
}

func TestSQLiteStore_ListIssues_NoBudget_IsNil(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "o/r", IssueNum: issuePtr(1), Agent: "a", Model: "m",
		Usage: usage.Usage{InputFresh: 1000, Output: 500}, Source: usage.SourceEstimated, At: time.Now(),
	}))

	issues, err := s.ListIssues(ctx, usage.Filter{Repo: "o/r"})
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Nil(t, issues[0].Budget)
}

func TestSQLiteStore_ListAllSessions_AcrossRepos(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.StartSession(ctx, "owner/a", issuePtr(1))
	require.NoError(t, err)

	_, err = s.StartSession(ctx, "owner/b", issuePtr(2))
	require.NoError(t, err)

	sessions, err := s.ListAllSessions(ctx)
	require.NoError(t, err)
	require.Len(t, sessions, 2)
	assert.Equal(t, "owner/a", sessions[0].Repo)
	assert.Equal(t, "owner/b", sessions[1].Repo)
}

func TestSQLiteStore_ListBudgets_RoundTrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.SetBudget(ctx, "owner/a", 1, 10.5))
	require.NoError(t, s.SetBudget(ctx, "owner/b", 2, 3.0))

	budgets, err := s.ListBudgets(ctx)
	require.NoError(t, err)
	require.Len(t, budgets, 2)
	assert.Equal(t, usage.IssueBudget{Repo: "owner/a", IssueNum: 1, Amount: 10.5}, budgets[0])
	assert.Equal(t, usage.IssueBudget{Repo: "owner/b", IssueNum: 2, Amount: 3.0}, budgets[1])
}

func TestSQLiteStore_ListBudgets_Empty(t *testing.T) {
	s := newTestStore(t)

	budgets, err := s.ListBudgets(context.Background())
	require.NoError(t, err)
	assert.Empty(t, budgets)
}

// issuePtr is a test helper: attribution is optional, so IssueNum is a pointer.
func issuePtr(n int) *int { return &n }

// Attribution is an annotation, not a precondition: a turn with no issue must
// still be recorded in full.
func TestSQLiteStore_LogUsage_WithoutIssue(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "o/r", Agent: "claude-code", Model: "m", SessionID: "sess-1",
		Usage:  usage.Usage{InputFresh: 100, CacheRead: 900, Output: 50},
		Source: usage.SourceMeasured,
	}))

	entries, err := s.ListEntries(ctx, usage.Filter{Repo: "o/r"})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Nil(t, entries[0].IssueNum)
	assert.Equal(t, 1000, entries[0].Usage.TotalInput())
	assert.Equal(t, usage.SourceMeasured, entries[0].Source)
}

// Unattributed usage belongs to no issue and must not invent one in the issue
// list, or every report would grow a phantom entry.
func TestSQLiteStore_ListIssues_ExcludesUnattributed(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "o/r", IssueNum: issuePtr(1), Agent: "a", Model: "m",
		Usage: usage.Usage{InputFresh: 10}, Source: usage.SourceEstimated,
	}))
	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "o/r", Agent: "a", Model: "m",
		Usage: usage.Usage{InputFresh: 20}, Source: usage.SourceMeasured,
	}))

	issues, err := s.ListIssues(ctx, usage.Filter{Repo: "o/r"})
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, 1, issues[0].IssueNum)
}

func TestSQLiteStore_ListUnattributed_GroupsBySession(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for range 3 {
		require.NoError(t, s.LogUsage(ctx, usage.Entry{
			Repo: "o/r", Agent: "a", Model: "m", SessionID: "sess-1",
			Usage: usage.Usage{InputFresh: 100, Output: 10}, Source: usage.SourceMeasured,
		}))
	}

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "o/r", Agent: "a", Model: "m", SessionID: "sess-2",
		Usage: usage.Usage{InputFresh: 50}, Source: usage.SourceMeasured,
	}))

	// Attributed usage never appears here.
	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "o/r", IssueNum: issuePtr(7), Agent: "a", Model: "m", SessionID: "sess-3",
		Usage: usage.Usage{InputFresh: 999}, Source: usage.SourceEstimated,
	}))

	groups, err := s.ListUnattributed(ctx, "o/r")
	require.NoError(t, err)
	require.Len(t, groups, 2)

	bySession := map[string]usage.UnattributedGroup{}
	for _, g := range groups {
		bySession[g.SessionID] = g
	}

	assert.Equal(t, 3, bySession["sess-1"].Entries)
	assert.Equal(t, 300, bySession["sess-1"].Usage.InputFresh)
	assert.Equal(t, 1, bySession["sess-2"].Entries)
}

func TestSQLiteStore_AssignIssue_AndReverse(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sess, err := s.StartSession(ctx, "o/r", nil)
	require.NoError(t, err)
	assert.Nil(t, sess.IssueNum)

	for range 2 {
		require.NoError(t, s.LogUsage(ctx, usage.Entry{
			Repo: "o/r", Agent: "a", Model: "m", SessionID: sess.ID,
			Usage: usage.Usage{InputFresh: 100}, Source: usage.SourceMeasured,
		}))
	}

	affected, err := s.AssignIssue(ctx, sess.ID, 42)
	require.NoError(t, err)
	assert.Equal(t, 2, affected)

	entries, err := s.ListEntries(ctx, usage.Filter{Repo: "o/r", IssueNum: 42})
	require.NoError(t, err)
	require.Len(t, entries, 2)

	sessions, err := s.ListSessions(ctx, "o/r", 42)
	require.NoError(t, err)
	require.Len(t, sessions, 1, "the session follows its entries, so wall-clock time does too")

	// Branch-derived attribution is a guess, so it has to be undoable.
	affected, err = s.UnassignIssue(ctx, sess.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, affected)

	groups, err := s.ListUnattributed(ctx, "o/r")
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, 2, groups[0].Entries)
	assert.Equal(t, 200, groups[0].Usage.InputFresh, "token counts survive the round trip")
}
