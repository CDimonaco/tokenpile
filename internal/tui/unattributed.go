package tui

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleUnattributedKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.unattributedCursor > 0 {
			m.unattributedCursor--
		}
	case "down", "j":
		if m.unattributedCursor < len(m.unattributed)-1 {
			m.unattributedCursor++
		}
	case "a", "enter":
		if len(m.unattributed) == 0 {
			return m, nil
		}

		// The suggestion is pre-filled, never applied: a branch-derived issue
		// is a guess, and a guess confirmed without being seen is worse than
		// no suggestion at all.
		m.assigning = true
		m.assignErr = nil
		m.assignInput = ""

		if s := m.unattributed[m.unattributedCursor].Suggested; s != nil {
			m.assignInput = strconv.Itoa(*s)
		}
	case "z":
		if m.lastAssigned == "" {
			return m, nil
		}

		return m, m.unassignCmd(m.lastAssigned)
	}

	return m, nil
}

func (m Model) handleAssignKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.assigning = false
		m.assignInput = ""

		return m, nil

	case "enter":
		n, err := strconv.Atoi(m.assignInput)
		if err != nil || n <= 0 {
			m.assignErr = errors.New("enter a positive issue number")

			return m, nil
		}

		m.assigning = false
		sessionID := m.unattributed[m.unattributedCursor].SessionID
		m.assignInput = ""

		return m, m.assignCmd(sessionID, n)

	case "backspace":
		if m.assignInput != "" {
			m.assignInput = m.assignInput[:len(m.assignInput)-1]
		}

		return m, nil
	}

	if r := msg.String(); len(r) == 1 && r[0] >= '0' && r[0] <= '9' {
		m.assignInput += r
	}

	return m, nil
}

// assignCmd and unassignCmd delegate to the same store methods the CLI uses,
// so the two surfaces cannot disagree about what assignment means.
func (m Model) assignCmd(sessionID string, issueNum int) tea.Cmd {
	return func() tea.Msg {
		if _, err := m.store.AssignIssue(context.Background(), sessionID, issueNum); err != nil {
			return assignmentDoneMsg{sessionID: sessionID, err: err}
		}

		return assignmentDoneMsg{sessionID: sessionID}
	}
}

func (m Model) unassignCmd(sessionID string) tea.Cmd {
	return func() tea.Msg {
		if _, err := m.store.UnassignIssue(context.Background(), sessionID); err != nil {
			return assignmentDoneMsg{sessionID: sessionID, undo: true, err: err}
		}

		return assignmentDoneMsg{sessionID: sessionID, undo: true}
	}
}

func (m Model) loadUnattributed() tea.Cmd {
	return func() tea.Msg {
		groups, err := m.store.ListUnattributed(context.Background(), "")
		if err != nil {
			return errMsg{err: err}
		}

		return unattributedLoadedMsg{groups: groups}
	}
}

func (m Model) viewUnattributed() string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s  %s\n", titleStyle.Render("tokenpile"), "unattributed usage")
	fmt.Fprintln(&b, dimStyle.Render("These tokens count toward no budget until they are assigned."))

	if len(m.unattributed) == 0 {
		fmt.Fprintln(&b, "\n"+dimStyle.Render("Nothing unattributed. Every measurement belongs to an issue."))

		return b.String()
	}

	fmt.Fprintln(&b, headerStyle.Render(fmt.Sprintf("%-18s %-22s %-19s %-7s %-9s %-9s %s",
		"Repo", "Branch", "When", "Entries", "Tokens", "Cost", "Suggested")))

	for i, g := range m.unattributed {
		line := fmt.Sprintf("%-18s %-22s %-19s %-7d %-9s %-9s %s",
			truncate(g.Repo, 18),
			truncate(branchLabel(g.Branch), 22),
			fmt.Sprintf("%s-%s", g.First.Local().Format("02 Jan 15:04"), g.Last.Local().Format("15:04")),
			g.Entries,
			fmt.Sprintf("%dk", g.Usage.TotalTokens()/1000),
			fmt.Sprintf("$%.2f", g.Cost),
			suggestionLabel(g.Suggested),
		)

		if i == m.unattributedCursor {
			fmt.Fprintln(&b, selectedStyle.Render(line))
		} else {
			fmt.Fprintln(&b, line)
		}
	}

	if m.assigning {
		fmt.Fprintf(&b, "\nAssign to issue #%s%s\n",
			m.assignInput, dimStyle.Render("_  (enter to confirm, esc to cancel)"))
	}

	if m.assignErr != nil {
		fmt.Fprintln(&b, errorStyle.Render("error: "+m.assignErr.Error()))
	}

	return b.String()
}

// branchLabel renders an unknown branch as a dash rather than blank: entries
// captured before the branch was stored have none, and that is a fact worth
// showing rather than an empty cell that reads as a rendering bug.
func branchLabel(branch string) string {
	if branch == "" {
		return "-"
	}

	return branch
}

func suggestionLabel(suggested *int) string {
	if suggested == nil {
		return "-"
	}

	return fmt.Sprintf("#%d", *suggested)
}
