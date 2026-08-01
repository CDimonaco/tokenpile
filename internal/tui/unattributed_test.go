package tui

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

func seedUnattributed(t *testing.T, s *store.SQLiteStore, sessionID, branch string) {
	t.Helper()

	require.NoError(t, s.LogUsage(context.Background(), usage.Entry{
		Repo:      "owner/repo",
		Agent:     "claude-code",
		Model:     "claude-sonnet-4-6",
		Usage:     usage.Usage{InputFresh: 1000, Output: 500},
		Branch:    branch,
		Source:    usage.SourceMeasured,
		SessionID: sessionID,
		At:        time.Now(),
	}))
}

func unattributedModel(groups []usage.UnattributedGroup) Model {
	return Model{activeView: viewUnattributed, unattributed: groups}
}

func issueNum(n int) *int { return &n }

func TestUnattributedView_ListsGroups(t *testing.T) {
	m := unattributedModel([]usage.UnattributedGroup{{
		Repo:      "owner/repo",
		SessionID: "sess-1",
		Branch:    "feat/42-thing",
		Entries:   3,
		Usage:     usage.Usage{InputFresh: 12000, Output: 3000},
		Cost:      0.42,
		First:     time.Now().Add(-time.Hour),
		Last:      time.Now(),
		Suggested: issueNum(42),
	}})

	v := stripANSI([]byte(m.View()))

	assert.Contains(t, v, "owner/repo")
	assert.Contains(t, v, "feat/42-thing")
	assert.Contains(t, v, "#42")
	assert.Contains(t, v, "$0.42")
}

func TestUnattributedView_EmptyStateSaysSo(t *testing.T) {
	v := stripANSI([]byte(unattributedModel(nil).View()))

	assert.Contains(t, v, "Nothing unattributed")
	assert.NotContains(t, v, "Suggested", "no table header without rows")
}

// A group whose branch implies nothing shows a dash, never a number: a wrong
// suggestion invites a careless confirmation.
func TestUnattributedView_NoSuggestionShowsDash(t *testing.T) {
	m := unattributedModel([]usage.UnattributedGroup{{
		Repo: "owner/repo", SessionID: "sess-1", Branch: "main", Entries: 1,
	}})

	v := stripANSI([]byte(m.View()))

	assert.Contains(t, v, "main")
	assert.NotContains(t, v, "#")
}

func TestUnattributedView_AssignPromptPreFillsSuggestion(t *testing.T) {
	m := unattributedModel([]usage.UnattributedGroup{{
		Repo: "owner/repo", SessionID: "sess-1", Branch: "feat/42-x", Suggested: issueNum(42),
	}})

	updated, _ := m.handleUnattributedKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updated.(Model)

	require.True(t, m.assigning)
	assert.Equal(t, "42", m.assignInput)
	assert.Contains(t, stripANSI([]byte(m.View())), "Assign to issue #42")

	// The pre-filled number is editable, not a fait accompli.
	updated, _ = m.handleAssignKey(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(Model)
	updated, _ = m.handleAssignKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("9")})
	m = updated.(Model)

	assert.Equal(t, "49", m.assignInput)
}

// Nothing is attributed until the assignment is confirmed.
func TestUnattributedView_OpeningThePromptAssignsNothing(t *testing.T) {
	s := newTUITestStore(t)
	seedUnattributed(t, s, "sess-1", "feat/42-x")

	m := newTUIModel(s)
	msg := m.loadUnattributed()()
	updated, _ := m.Update(msg)
	m = updated.(Model)
	m.activeView = viewUnattributed

	updated, _ = m.handleUnattributedKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updated.(Model)
	require.True(t, m.assigning)

	groups, err := s.ListUnattributed(context.Background(), "owner/repo")
	require.NoError(t, err)
	assert.Len(t, groups, 1, "still unattributed")
}

func TestUnattributedView_AssigningRemovesTheGroup(t *testing.T) {
	s := newTUITestStore(t)
	seedUnattributed(t, s, "sess-1", "feat/42-x")

	m := newTUIModel(s)
	updated, _ := m.Update(m.loadUnattributed()())
	m = updated.(Model)
	m.activeView = viewUnattributed
	require.Len(t, m.unattributed, 1)

	// Open the prompt on the suggestion and confirm it.
	updated, _ = m.handleUnattributedKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updated.(Model)
	updated, cmd := m.handleAssignKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	require.NotNil(t, cmd)

	done, ok := cmd().(assignmentDoneMsg)
	require.True(t, ok)
	require.NoError(t, done.err)

	updated, _ = m.Update(done)
	m = updated.(Model)
	updated, _ = m.Update(m.loadUnattributed()())
	m = updated.(Model)

	assert.Empty(t, m.unattributed, "the group is gone from the list")

	entries, err := s.ListEntries(context.Background(), usage.Filter{Repo: "owner/repo", IssueNum: 42})
	require.NoError(t, err)
	assert.Len(t, entries, 1, "its usage moved to the suggested issue")
}

// Assignment is a guess when it comes from a branch, so it has to be undoable
// without leaving the view.
func TestUnattributedView_UndoReturnsTheGroup(t *testing.T) {
	s := newTUITestStore(t)
	seedUnattributed(t, s, "sess-1", "feat/42-x")

	m := newTUIModel(s)
	updated, _ := m.Update(m.loadUnattributed()())
	m = updated.(Model)
	m.activeView = viewUnattributed

	updated, _ = m.Update(m.assignCmd("sess-1", 42)())
	m = updated.(Model)
	require.Equal(t, "sess-1", m.lastAssigned)

	updated, _ = m.handleUnattributedKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	m = updated.(Model)

	updated, _ = m.Update(m.unassignCmd("sess-1")())
	m = updated.(Model)
	updated, _ = m.Update(m.loadUnattributed()())
	m = updated.(Model)

	require.Len(t, m.unattributed, 1)
	assert.Equal(t, 1500, m.unattributed[0].Usage.TotalTokens(), "token counts survive the round trip")
	assert.Empty(t, m.lastAssigned, "nothing left to undo")
}

func TestIssueList_IndicatesUnattributedUsage(t *testing.T) {
	m := Model{
		activeView: viewList,
		issues:     []usage.TrackedIssue{{IssueNum: 7, Repo: "owner/repo"}},
		unattributed: []usage.UnattributedGroup{
			{Repo: "owner/repo", SessionID: "sess-1"},
		},
	}

	v := stripANSI([]byte(m.View()))

	assert.Contains(t, v, "1 unattributed session(s)")
	assert.Contains(t, v, "press u to review")
}

func TestIssueList_NoIndicatorWhenNothingUnattributed(t *testing.T) {
	m := Model{
		activeView: viewList,
		issues:     []usage.TrackedIssue{{IssueNum: 7, Repo: "owner/repo"}},
	}

	// The footer still advertises the view; the list must not claim there is
	// anything in it.
	assert.NotContains(t, stripANSI([]byte(m.View())), "press u to review")
}

// The prompt swallows every key while open: typing an issue number containing
// no digits must not quit the program.
func TestAssignPrompt_QDoesNotQuit(t *testing.T) {
	m := unattributedModel([]usage.UnattributedGroup{{Repo: "owner/repo", SessionID: "sess-1"}})
	m.assigning = true

	updated, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updated.(Model)

	assert.Nil(t, cmd)
	assert.True(t, m.assigning)
	assert.Empty(t, m.assignInput, "a letter is not accepted into a number field")
}

func TestAssignPrompt_EscCancels(t *testing.T) {
	m := unattributedModel([]usage.UnattributedGroup{{Repo: "owner/repo", SessionID: "sess-1"}})
	m.assigning = true
	m.assignInput = "42"

	updated, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	assert.False(t, m.assigning)
	assert.Empty(t, m.assignInput)
}
