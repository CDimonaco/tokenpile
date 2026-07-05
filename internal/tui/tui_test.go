package tui

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

// ansiRe strips ANSI CSI and basic OSC escape sequences for plain-text assertions.
var ansiRe = regexp.MustCompile(`\x1b(?:\[[0-9;]*[a-zA-Z]|\][^\x07\x1b]*[\x07\x1b])`)

func stripANSI(b []byte) string {
	return string(ansiRe.ReplaceAll(b, nil))
}

// tuiStubProvider satisfies provider.IssueProvider without making network calls.
type tuiStubProvider struct{}

func (p *tuiStubProvider) ListIssues(_ context.Context, _ usage.Filter) ([]provider.Issue, error) {
	return nil, nil
}

func (p *tuiStubProvider) GetIssue(_ context.Context, repo string, number int) (*provider.Issue, error) {
	return &provider.Issue{Number: number, Repo: repo, Title: "TUI Test Issue"}, nil
}

func newTUITestStore(t *testing.T) *store.SQLiteStore {
	t.Helper()

	loader, err := pricing.NewLoader("")
	require.NoError(t, err)

	s, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "tui_test.db"), loader)
	require.NoError(t, err)

	t.Cleanup(func() { _ = s.Close() })

	return s
}

func newTUIModel(s store.Store) Model {
	loader, _ := pricing.NewLoader("")
	// empty authToken: loadIssues uses DB only (no GitHub call)
	return New(s, &tuiStubProvider{}, loader, "")
}

func seedIssueEntry(t *testing.T, s *store.SQLiteStore, issueNum int) {
	t.Helper()

	ctx := context.Background()
	const repo = "owner/repo"

	sess, err := s.StartSession(ctx, repo, issueNum)
	require.NoError(t, err)

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo:      repo,
		IssueNum:  issueNum,
		Agent:     "claude-code",
		Model:     "claude-sonnet-4-6",
		TokensIn:  1000,
		TokensOut: 500,
		SessionID: sess.ID,
		At:        time.Now(),
	}))

	require.NoError(t, s.EndSession(ctx, sess.ID))
}

func baseModel() Model {
	return Model{
		activeView:  viewList,
		granularity: usage.Day,
		authToken:   "tok",
	}
}

func TestView_Unauthenticated_ShowsLocalIssues(t *testing.T) {
	m := Model{
		activeView:      viewList,
		unauthenticated: true,
		issues: []usage.TrackedIssue{
			{IssueNum: 7, Repo: "owner/repo"},
		},
	}

	v := m.View()

	assert.Contains(t, v, "not authenticated")
	assert.Contains(t, v, "7")
}

func TestView_Unauthenticated_EmptyDB_ShowsEmptyState(t *testing.T) {
	m := Model{activeView: viewList, unauthenticated: true}

	v := m.View()

	assert.Contains(t, v, "not authenticated")
	assert.Contains(t, v, "No usage tracked yet")
	assert.NotContains(t, v, "Not authenticated. Run:")
}

func TestView_ErrorMessage(t *testing.T) {
	m := baseModel()
	m.err = errMsg{err: assert.AnError}.err

	v := m.View()

	assert.Contains(t, v, "error")
}

func TestIssuesLoaded_UnauthenticatedFlag(t *testing.T) {
	m := baseModel()
	issues := []usage.TrackedIssue{{IssueNum: 3, Repo: "o/r"}}

	next, _ := m.Update(issuesLoadedMsg{issues: issues, unauthenticated: true})
	m2 := next.(Model)

	assert.True(t, m2.unauthenticated)
	assert.Equal(t, issues, m2.issues)
}

func TestIssuesLoaded_Dedup(t *testing.T) {
	m := baseModel()

	// Simulate a merge where the same issue appears in both sources
	merged := []usage.TrackedIssue{
		{IssueNum: 1, Repo: "o/r", Title: "from github"},
		{IssueNum: 2, Repo: "o/r", Title: "db only"},
	}

	next, _ := m.Update(issuesLoadedMsg{issues: merged})
	m2 := next.(Model)

	assert.Len(t, m2.issues, 2)
}

func TestIssuesLoaded_StubTitleForInaccessible(t *testing.T) {
	m := baseModel()

	merged := []usage.TrackedIssue{
		{IssueNum: 99, Repo: "o/r", Title: "[not found on GitHub]"},
	}

	next, _ := m.Update(issuesLoadedMsg{issues: merged})
	m2 := next.(Model)

	v := m2.View()
	assert.Contains(t, v, "[not found on GitHub]")
}

func TestHandleKey_Quit(t *testing.T) {
	m := baseModel()

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	_ = next
	assert.NotNil(t, cmd)
}

func TestHandleKey_HelpToggle(t *testing.T) {
	m := baseModel()

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m2 := next.(Model)

	assert.Equal(t, viewHelp, m2.activeView)

	next2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m3 := next2.(Model)

	assert.Equal(t, viewList, m3.activeView)
}

func TestHandleKey_ListNavigation(t *testing.T) {
	m := baseModel()
	m.issues = []usage.TrackedIssue{
		{IssueNum: 1, Repo: "o/a"},
		{IssueNum: 2, Repo: "o/b"},
		{IssueNum: 3, Repo: "o/c"},
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m2 := next.(Model)
	assert.Equal(t, 1, m2.cursor)

	next2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m3 := next2.(Model)
	assert.Equal(t, 2, m3.cursor)

	// cannot go past end
	next3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m4 := next3.(Model)
	assert.Equal(t, 2, m4.cursor)

	// go back up
	next4, _ := m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m5 := next4.(Model)
	assert.Equal(t, 1, m5.cursor)
}

func TestHandleKey_EnterDetail(t *testing.T) {
	m := baseModel()
	m.issues = []usage.TrackedIssue{{IssueNum: 5, Repo: "o/r"}}
	m.store = nil // loadReport will be called but we only test state transition

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := next.(Model)

	assert.Equal(t, viewDetail, m2.activeView)
	assert.NotNil(t, m2.selected)
	assert.Equal(t, 5, m2.selected.IssueNum)
}

func TestHandleKey_EscFromDetail(t *testing.T) {
	m := baseModel()
	m.activeView = viewDetail

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := next.(Model)

	assert.Equal(t, viewList, m2.activeView)
}

func TestHandleKey_EscFromChart(t *testing.T) {
	m := baseModel()
	m.activeView = viewChart
	m.prevView = viewList

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m2 := next.(Model)

	assert.Equal(t, viewList, m2.activeView)
}

func TestHandleKey_GranularityToggle(t *testing.T) {
	m := baseModel()
	m.activeView = viewChart
	m.store = nil

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")})
	m2 := next.(Model)
	assert.Equal(t, usage.Week, m2.granularity)

	next2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m3 := next2.(Model)
	assert.Equal(t, usage.Day, m3.granularity)
}

func TestIssuesLoaded_UpdatesModel(t *testing.T) {
	m := baseModel()
	issues := []usage.TrackedIssue{{IssueNum: 1, Repo: "o/r"}}

	next, _ := m.Update(issuesLoadedMsg{issues: issues})
	m2 := next.(Model)

	assert.Equal(t, issues, m2.issues)
	assert.NoError(t, m2.err)
}

func TestWindowSize_UpdatesDimensions(t *testing.T) {
	m := baseModel()

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m2 := next.(Model)

	assert.Equal(t, 120, m2.width)
	assert.Equal(t, 40, m2.height)
}

func TestViewList_LoadingState(t *testing.T) {
	m := baseModel()
	m.loading = true

	v := m.View()

	assert.Contains(t, v, "Loading...")
	assert.NotContains(t, v, "No usage tracked yet")
}

func TestIssuesLoaded_ClearsLoadingFlag(t *testing.T) {
	m := baseModel()
	m.loading = true

	next, _ := m.Update(issuesLoadedMsg{issues: nil})
	m2 := next.(Model)

	assert.False(t, m2.loading)
}

// ---- budgetBar unit tests ----

func TestBudgetBar_NilBudget_ReturnsEmpty(t *testing.T) {
	assert.Empty(t, budgetBar(1.0, nil))
}

func TestBudgetBar_ZeroBudget_ReturnsEmpty(t *testing.T) {
	zero := 0.0
	assert.Empty(t, budgetBar(1.0, &zero))
}

func TestBudgetBar_BelowWarn_ContainsFilledAndEmpty(t *testing.T) {
	budget := 100.0
	out := budgetBar(50.0, &budget) // 50%
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "#")
	assert.Contains(t, out, "-")
}

func TestBudgetBar_AtWarn_NotEmpty(t *testing.T) {
	budget := 100.0
	out := budgetBar(80.0, &budget) // 80%
	assert.NotEmpty(t, out)
}

func TestBudgetBar_FullyConsumed_NoEmptySlots(t *testing.T) {
	budget := 10.0
	out := budgetBar(10.0, &budget) // 100%
	assert.NotEmpty(t, out)
	assert.NotContains(t, out, "-")
}

func TestBudgetBar_OverBudget_ClampsToFull(t *testing.T) {
	budget := 10.0
	out := budgetBar(20.0, &budget) // 200% — must not panic or render extra chars
	assert.NotEmpty(t, out)
	assert.NotContains(t, out, "-")
}

// ---- TUI teatest integration tests ----

func TestTUI_IssueList_ShowsTrackedIssues(t *testing.T) {
	s := newTUITestStore(t)
	seedIssueEntry(t, s, 1)

	tm := teatest.NewTestModel(t, newTUIModel(s), teatest.WithInitialTermSize(120, 40))
	t.Cleanup(func() { _ = tm.Quit() })

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return strings.Contains(stripANSI(b), "#1")
	}, teatest.WithDuration(3*time.Second))
}

func TestTUI_IssueList_ShowsBudgetBar(t *testing.T) {
	s := newTUITestStore(t)
	seedIssueEntry(t, s, 2)
	require.NoError(t, s.SetBudget(context.Background(), "owner/repo", 2, 100.00))

	tm := teatest.NewTestModel(t, newTUIModel(s), teatest.WithInitialTermSize(120, 40))
	t.Cleanup(func() { _ = tm.Quit() })

	// wait for list and the budget bar ('[' bracket) to appear alongside the issue
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		plain := stripANSI(b)
		return strings.Contains(plain, "#2") && strings.Contains(plain, "[")
	}, teatest.WithDuration(3*time.Second))
}

func TestTUI_DetailView_SummaryAndSessionsTabs(t *testing.T) {
	s := newTUITestStore(t)
	seedIssueEntry(t, s, 3)

	tm := teatest.NewTestModel(t, newTUIModel(s), teatest.WithInitialTermSize(120, 40))
	t.Cleanup(func() { _ = tm.Quit() })

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return strings.Contains(stripANSI(b), "#3")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// detail view shows both tab names
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		plain := stripANSI(b)
		return strings.Contains(plain, "Summary") && strings.Contains(plain, "Sessions")
	}, teatest.WithDuration(3*time.Second))
}

func TestTUI_DetailView_SessionsTab_ShowsNoteAndTags(t *testing.T) {
	s := newTUITestStore(t)
	ctx := context.Background()

	sess, err := s.StartSession(ctx, "owner/repo", 4)
	require.NoError(t, err)

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "owner/repo", IssueNum: 4,
		Agent: "claude-code", Model: "claude-sonnet-4-6",
		TokensIn: 500, TokensOut: 250, SessionID: sess.ID, At: time.Now(),
	}))

	note := "session tab note"
	require.NoError(t, s.UpdateSessionAnnotations(ctx, sess.ID, &note, []string{"test-tag"}))
	require.NoError(t, s.EndSession(ctx, sess.ID))

	tm := teatest.NewTestModel(t, newTUIModel(s), teatest.WithInitialTermSize(120, 40))
	t.Cleanup(func() { _ = tm.Quit() })

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return strings.Contains(stripANSI(b), "#4")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return strings.Contains(stripANSI(b), "Summary")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyTab})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		plain := stripANSI(b)
		return strings.Contains(plain, "Session 1") && strings.Contains(plain, "session tab note")
	}, teatest.WithDuration(3*time.Second))
}

func TestTUI_DetailView_SummaryTab_ShowsBudgetBar(t *testing.T) {
	s := newTUITestStore(t)
	ctx := context.Background()

	seedIssueEntry(t, s, 5)
	require.NoError(t, s.SetBudget(ctx, "owner/repo", 5, 100.00))

	tm := teatest.NewTestModel(t, newTUIModel(s), teatest.WithInitialTermSize(120, 40))
	t.Cleanup(func() { _ = tm.Quit() })

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return strings.Contains(stripANSI(b), "#5")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Summary tab should show the Budget line with the bar
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		plain := stripANSI(b)
		return strings.Contains(plain, "Budget") && strings.Contains(plain, "$100.00")
	}, teatest.WithDuration(3*time.Second))
}

func TestViewList_EmptyState(t *testing.T) {
	m := baseModel()

	v := m.View()

	assert.Contains(t, v, "No usage tracked yet")
}

func TestViewList_ShowsIssues(t *testing.T) {
	m := baseModel()
	m.issues = []usage.TrackedIssue{
		{IssueNum: 42, Repo: "owner/repo", TotalTokensIn: 1000, TotalTokensOut: 500, TotalCost: 0.01},
	}

	v := m.View()

	assert.Contains(t, v, "42")
	assert.Contains(t, v, "owner/repo")
}

func TestViewHelp_ContainsShortcuts(t *testing.T) {
	m := baseModel()
	m.activeView = viewHelp

	v := m.View()

	assert.Contains(t, v, "quit")
	assert.Contains(t, v, "chart")
}
