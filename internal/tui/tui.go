package tui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cdimonaco/tokenpile/internal/domain"
	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/store"
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

	activeView  view
	prevView    view
	issues      []domain.TrackedIssue
	selected    *domain.TrackedIssue
	report      *domain.Report
	chartPoints []domain.UsagePoint
	chartScope  string

	granularity domain.Granularity
	filterAgent string
	filterModel string
	filterFrom  *time.Time
	filterTo    *time.Time

	width   int
	height  int
	cursor  int
	err     error
	authErr bool
}

type (
	issuesLoadedMsg struct{ issues []domain.TrackedIssue }
	reportLoadedMsg struct{ report *domain.Report }
	chartLoadedMsg  struct{ points []domain.UsagePoint }
	errMsg          struct{ err error }
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("236")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Underline(true)
)

func New(s store.Store, ip provider.IssueProvider, p *pricing.Loader, authToken string) Model {
	return Model{
		store:         s,
		issueProvider: ip,
		pricer:        p,
		authToken:     authToken,
		activeView:    viewList,
		granularity:   domain.GranularityDay,
	}
}

func (m Model) Init() tea.Cmd {
	if m.authToken == "" {
		m.authErr = true //nolint:govet

		return nil
	}

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
		m.granularity = domain.GranularityDay

		return m, m.loadChart(m.chartIssueNum())
	case "w":
		m.granularity = domain.GranularityWeek

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
	if m.authErr {
		return errorStyle.Render("Not authenticated. Run: tokenpile auth login --provider github")
	}

	if m.err != nil {
		return errorStyle.Render("Error: " + m.err.Error())
	}

	switch m.activeView {
	case viewList:
		return m.viewIssueList()
	case viewDetail:
		return m.viewIssueDetail()
	case viewChart:
		return m.viewChart()
	case viewHelp:
		return m.viewHelp()
	}

	return ""
}

func (m Model) viewIssueList() string {
	var b strings.Builder

	fmt.Fprintln(&b, titleStyle.Render("tokenpile"))
	fmt.Fprintln(&b, headerStyle.Render(fmt.Sprintf("%-8s %-20s %-12s %-12s %s", "Issue", "Repo", "Tokens", "Cost", "Time")))

	if len(m.issues) == 0 {
		fmt.Fprintln(&b, dimStyle.Render("No usage tracked yet. Run tokenpile log to get started."))

		return b.String()
	}

	for i, issue := range m.issues {
		line := fmt.Sprintf("#%-7d %-20s %-12s %-12s %s",
			issue.IssueNum,
			truncate(issue.Repo, 20),
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

	fmt.Fprintln(&b, dimStyle.Render("\nenter: detail  c: chart  ?: help  q: quit"))

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

	fmt.Fprintln(&b, dimStyle.Render("\nc: chart  esc: back  ?: help  q: quit"))

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
	fmt.Fprintln(&b, dimStyle.Render("d: day  w: week  esc: back  ?: help  q: quit"))

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

		issues, err := m.store.ListIssues(ctx, domain.IssueFilter{})
		if err != nil {
			slog.Error("load issues", "err", err)
			return errMsg{err: err}
		}

		return issuesLoadedMsg{issues: issues}
	}
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

		filter := domain.UsageOverTimeFilter{
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
