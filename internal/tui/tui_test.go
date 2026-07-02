package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/cdimonaco/tokenpile/internal/domain"
)

func baseModel() Model {
	return Model{
		activeView:  viewList,
		granularity: domain.GranularityDay,
		authToken:   "tok",
	}
}

func TestView_AuthError_WhenNoToken(t *testing.T) {
	m := Model{activeView: viewList}
	m.authErr = true

	v := m.View()

	assert.Contains(t, v, "Not authenticated")
}

func TestView_ErrorMessage(t *testing.T) {
	m := baseModel()
	m.err = errMsg{err: assert.AnError}.err

	v := m.View()

	assert.Contains(t, v, "Error")
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
	m.issues = []domain.TrackedIssue{
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
	m.issues = []domain.TrackedIssue{{IssueNum: 5, Repo: "o/r"}}
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
	assert.Equal(t, domain.GranularityWeek, m2.granularity)

	next2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m3 := next2.(Model)
	assert.Equal(t, domain.GranularityDay, m3.granularity)
}

func TestIssuesLoaded_UpdatesModel(t *testing.T) {
	m := baseModel()
	issues := []domain.TrackedIssue{{IssueNum: 1, Repo: "o/r"}}

	next, _ := m.Update(issuesLoadedMsg{issues: issues})
	m2 := next.(Model)

	assert.Equal(t, issues, m2.issues)
	assert.Nil(t, m2.err)
}

func TestWindowSize_UpdatesDimensions(t *testing.T) {
	m := baseModel()

	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m2 := next.(Model)

	assert.Equal(t, 120, m2.width)
	assert.Equal(t, 40, m2.height)
}

func TestViewList_EmptyState(t *testing.T) {
	m := baseModel()

	v := m.View()

	assert.Contains(t, v, "No usage tracked yet")
}

func TestViewList_ShowsIssues(t *testing.T) {
	m := baseModel()
	m.issues = []domain.TrackedIssue{
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
