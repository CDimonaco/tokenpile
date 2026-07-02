package tui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

type view int

const (
	viewList view = iota
	viewDetail
	viewChart
	viewHelp
)

type Model struct {
	store         store.Store
	issueProvider provider.IssueProvider
	pricer        *pricing.Loader
	authToken     string

	activeView    view
	prevView      view
	issues        []usage.TrackedIssue
	selected      *usage.TrackedIssue
	report        *usage.Report
	chartPoints   []usage.Point
	chartScope    string
	unauthenticated bool

	granularity usage.Granularity
	filterAgent string
	filterModel string
	filterFrom  *time.Time
	filterTo    *time.Time

	width  int
	height int
	cursor int
	err    error
}

type (
	issuesLoadedMsg struct {
		issues          []usage.TrackedIssue
		unauthenticated bool
	}
	reportLoadedMsg struct{ report *usage.Report }
	chartLoadedMsg  struct{ points []usage.Point }
	errMsg          struct{ err error }
)

const contentWidth = 90

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("236")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Underline(true)
	keyStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	footerSep     = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

func New(s store.Store, ip provider.IssueProvider, p *pricing.Loader, authToken string) Model {
	return Model{
		store:         s,
		issueProvider: ip,
		pricer:        p,
		authToken:     authToken,
		activeView:    viewList,
		granularity:   usage.Day,
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadIssues()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		return m, nil

	case issuesLoadedMsg:
		m.issues = msg.issues
		m.unauthenticated = msg.unauthenticated
		m.err = nil

		return m, nil

	case reportLoadedMsg:
		m.report = msg.report

		return m, nil

	case chartLoadedMsg:
		m.chartPoints = msg.points

		return m, nil

	case errMsg:
		m.err = msg.err

		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "?":
		if m.activeView == viewHelp {
			m.activeView = m.prevView
		} else {
			m.prevView = m.activeView
			m.activeView = viewHelp
		}

		return m, nil

	case "esc":
		switch m.activeView {
		case viewDetail:
			m.activeView = viewList

			return m, nil
		case viewChart:
			m.activeView = m.prevView

			return m, nil
		}
	}

	switch m.activeView {
	case viewList:
		return m.handleListKey(msg)
	case viewDetail:
		return m.handleDetailKey(msg)
	case viewChart:
		return m.handleChartKey(msg)
	}

	return m, nil
}

func (m Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.issues)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.issues) > 0 {
			issue := m.issues[m.cursor]
			m.selected = &issue
			m.activeView = viewDetail

			return m, m.loadReport(issue.Repo, issue.IssueNum)
		}
	case "c":
		m.prevView = viewList
		m.activeView = viewChart

		return m, m.loadChart(nil)
	}

	return m, nil
}

func (m Model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		if m.selected != nil {
			m.prevView = viewDetail
			m.activeView = viewChart

			return m, m.loadChart(&m.selected.IssueNum)
		}
	}

	return m, nil
}

func (m Model) handleChartKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "d":
		m.granularity = usage.Day

		return m, m.loadChart(m.chartIssueNum())
	case "w":
		m.granularity = usage.Week

		return m, m.loadChart(m.chartIssueNum())
	}

	return m, nil
}

func (m *Model) chartIssueNum() *int {
	if m.prevView == viewDetail && m.selected != nil {
		return &m.selected.IssueNum
	}

	return nil
}

func (m Model) View() string {
	var body string

	if m.err != nil {
		body = errorStyle.Render("error: " + m.err.Error())
	} else {
		switch m.activeView {
		case viewList:
			body = m.viewIssueList()
		case viewDetail:
			body = m.viewIssueDetail()
		case viewChart:
			body = m.viewChart()
		case viewHelp:
			body = m.viewHelp()
		}
	}

	w := contentWidth
	if m.width > 0 && m.width < w {
		w = m.width
	}

	content := lipgloss.NewStyle().Width(w).MarginTop(3).Render(body + "\n" + m.renderFooter())

	if m.width == 0 || m.height == 0 {
		return content
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)
}

func (m Model) renderFooter() string {
	sep := footerSep.Render(strings.Repeat("─", contentWidth))

	var keys []string

	switch m.activeView {
	case viewList:
		keys = []string{
			keyStyle.Render("j/k") + " navigate",
			keyStyle.Render("enter") + " detail",
			keyStyle.Render("c") + " chart",
			keyStyle.Render("?") + " help",
			keyStyle.Render("q") + " quit",
		}
	case viewDetail:
		keys = []string{
			keyStyle.Render("c") + " chart",
			keyStyle.Render("esc") + " back",
			keyStyle.Render("?") + " help",
			keyStyle.Render("q") + " quit",
		}
	case viewChart:
		keys = []string{
			keyStyle.Render("d") + " day",
			keyStyle.Render("w") + " week",
			keyStyle.Render("esc") + " back",
			keyStyle.Render("?") + " help",
			keyStyle.Render("q") + " quit",
		}
	case viewHelp:
		keys = []string{
			keyStyle.Render("?") + " close help",
			keyStyle.Render("q") + " quit",
		}
	}

	return sep + "\n" + dimStyle.Render(strings.Join(keys, "   "))
}

func (m Model) viewIssueList() string {
	var b strings.Builder

	fmt.Fprintln(&b, titleStyle.Render("tokenpile"))

	if m.unauthenticated {
		fmt.Fprintln(&b, dimStyle.Render("(not authenticated — showing local issues only)"))
	}

	fmt.Fprintln(&b, headerStyle.Render(fmt.Sprintf("%-8s %-18s %-26s %-10s %-10s %s", "Issue", "Repo", "Title", "Tokens", "Cost", "Time")))

	if len(m.issues) == 0 {
		fmt.Fprintln(&b, dimStyle.Render("No usage tracked yet. Run tokenpile log to get started."))

		return b.String()
	}

	for i, issue := range m.issues {
		title := issue.Title
		if title == "" {
			title = "-"
		}

		line := fmt.Sprintf("#%-7d %-18s %-26s %-10s %-10s %s",
			issue.IssueNum,
			truncate(issue.Repo, 18),
			truncate(title, 26),
			fmt.Sprintf("%dk", (issue.TotalTokensIn+issue.TotalTokensOut)/1000),
			fmt.Sprintf("$%.2f", issue.TotalCost),
			formatDuration(issue.TotalTime),
		)

		if i == m.cursor {
			fmt.Fprintln(&b, selectedStyle.Render(line))
		} else {
			fmt.Fprintln(&b, line)
		}
	}

	return b.String()
}

func (m Model) viewIssueDetail() string {
	if m.selected == nil {
		return ""
	}

	var b strings.Builder

	fmt.Fprintf(&b, "%s Issue #%d — %s\n\n", titleStyle.Render("tokenpile"), m.selected.IssueNum, m.selected.Repo)

	if m.report == nil {
		fmt.Fprintln(&b, "Loading...")

		return b.String()
	}

	fmt.Fprintln(&b, headerStyle.Render(fmt.Sprintf("%-16s %-24s %-8s %-12s %-12s %s", "Agent", "Model", "Calls", "In", "Out", "Cost")))

	for _, row := range m.report.Rows {
		fmt.Fprintf(&b, "%-16s %-24s %-8d %-12s %-12s $%.4f\n",
			truncate(row.Agent, 16),
			truncate(row.Model, 24),
			row.Calls,
			fmt.Sprintf("%dk", row.TokensIn/1000),
			fmt.Sprintf("%dk", row.TokensOut/1000),
			row.Cost,
		)
	}

	fmt.Fprintf(&b, "\n%s  in: %dk  out: %dk  cost: $%.4f  time: %s\n",
		headerStyle.Render("Total"),
		m.report.TotalTokensIn/1000,
		m.report.TotalTokensOut/1000,
		m.report.TotalCost,
		formatDuration(m.report.TotalTime),
	)

	return b.String()
}

func (m Model) viewChart() string {
	var b strings.Builder

	scope := "all issues"
	if m.prevView == viewDetail && m.selected != nil {
		scope = fmt.Sprintf("issue #%d", m.selected.IssueNum)
	}

	fmt.Fprintf(&b, "%s Token usage — %s  [%s]\n\n",
		titleStyle.Render("tokenpile"),
		scope,
		m.granularity,
	)

	if len(m.chartPoints) == 0 {
		fmt.Fprintln(&b, dimStyle.Render("No data for selected range."))
	} else {
		maxTokens := 0

		for _, p := range m.chartPoints {
			if total := p.TokensIn + p.TokensOut; total > maxTokens {
				maxTokens = total
			}
		}

		barWidth := 40

		for _, p := range m.chartPoints {
			total := p.TokensIn + p.TokensOut
			filled := 0

			if maxTokens > 0 {
				filled = total * barWidth / maxTokens
			}

			bar := strings.Repeat("#", filled) + strings.Repeat("-", barWidth-filled)
			fmt.Fprintf(&b, "%s |%s| %dk\n", p.Date.Format("2006-01-02"), bar, total/1000)
		}
	}

	var totalIn, totalOut int
	var totalCost float64

	for _, p := range m.chartPoints {
		totalIn += p.TokensIn
		totalOut += p.TokensOut
		totalCost += p.Cost
	}

	fmt.Fprintf(&b, "\nTotal: in %dk  out %dk  cost $%.4f\n", totalIn/1000, totalOut/1000, totalCost)

	return b.String()
}

func (m Model) viewHelp() string {
	return `tokenpile — keyboard shortcuts

  Navigation
    j / down    move down
    k / up      move up
    enter       open issue detail
    esc         go back

  Views
    c           open chart view
    ?           toggle this help

  Chart
    d           day granularity
    w           week granularity

  General
    q / ctrl+c  quit
`
}

func (m Model) loadIssues() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		dbIssues, err := m.store.ListIssues(ctx, usage.Filter{})
		if err != nil {
			slog.Error("load issues", "err", err)
			return errMsg{err: err}
		}

		if m.authToken == "" {
			return issuesLoadedMsg{issues: dbIssues, unauthenticated: true}
		}

		// Build dedup index keyed by "repo#issueNum"
		result := make([]usage.TrackedIssue, len(dbIssues))
		copy(result, dbIssues)
		inResult := make(map[string]int, len(result))
		for i, ti := range result {
			inResult[issueKey(ti.Repo, ti.IssueNum)] = i
		}

		repo, repoErr := provider.InferRepo()
		if repoErr == nil {
			ghIssues, ghErr := m.issueProvider.ListIssues(ctx, usage.Filter{
				Repo:     repo,
				Assignee: "@me",
			})
			if ghErr == nil {
				for _, gi := range ghIssues {
					k := issueKey(gi.Repo, gi.Number)
					if idx, ok := inResult[k]; ok {
						result[idx].Title = gi.Title
					} else {
						idx = len(result)
						result = append(result, usage.TrackedIssue{
							IssueNum: gi.Number,
							Repo:     gi.Repo,
							Title:    gi.Title,
						})
						inResult[k] = idx
					}
				}
			}
		}

		// Enrich DB issues still missing a title via GetIssue
		for i := range result {
			if result[i].Title != "" {
				continue
			}

			gi, getErr := m.issueProvider.GetIssue(ctx, result[i].Repo, result[i].IssueNum)
			if getErr != nil {
				result[i].Title = "[not found on GitHub]"
			} else {
				result[i].Title = gi.Title
			}
		}

		return issuesLoadedMsg{issues: result}
	}
}

func issueKey(repo string, num int) string {
	return fmt.Sprintf("%s#%d", repo, num)
}

func (m Model) loadReport(repo string, issueNum int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		report, err := m.store.GetReport(ctx, repo, issueNum)
		if err != nil {
			return errMsg{err: err}
		}

		return reportLoadedMsg{report: report}
	}
}

func (m Model) loadChart(issueNum *int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		filter := usage.OverTimeFilter{
			Granularity: m.granularity,
			Agent:       m.filterAgent,
			Model:       m.filterModel,
			From:        m.filterFrom,
			To:          m.filterTo,
		}

		if issueNum != nil {
			filter.IssueNum = issueNum
		}

		points, err := m.store.ListUsageOverTime(ctx, filter)
		if err != nil {
			return errMsg{err: err}
		}

		return chartLoadedMsg{points: points}
	}
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}

	h := int(d.Hours())
	m := int(d.Minutes()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}

	return fmt.Sprintf("%dm", m)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[:n-1] + "."
}
