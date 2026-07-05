package usage

import "time"

type Entry struct {
	ID          string
	Repo        string
	IssueNum    int
	Agent       string
	Model       string
	TokensIn    int
	TokensOut   int
	SessionID   string
	At          time.Time
	IssueTitle  string
	IssueLabels []string
}

type Session struct {
	ID             string
	Repo           string
	IssueNum       int
	StartedAt      time.Time
	EndedAt        *time.Time
	LastActivityAt time.Time
	Note           string
	Tags           []string
}

// Filter scopes queries over usage data.
type Filter struct {
	Repo     string
	IssueNum int
	State    string
	Assignee string
	Agent    string
	Model    string
	From     *time.Time
	To       *time.Time
}

type ReportRow struct {
	Agent     string
	Model     string
	Calls     int
	TokensIn  int
	TokensOut int
	Cost      float64
}

type Report struct {
	IssueNum       int
	Repo           string
	Rows           []ReportRow
	TotalTokensIn  int
	TotalTokensOut int
	TotalCost      float64
	TotalTime      time.Duration
}

type IssueCache struct {
	Repo      string
	IssueNum  int
	Title     string
	Labels    []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TrackedIssue struct {
	IssueNum       int
	Repo           string
	Title          string
	Labels         []string
	TotalTokensIn  int
	TotalTokensOut int
	TotalCost      float64
	TotalTime      time.Duration
	Budget         *float64
}

type TrackedIssueRef struct {
	Repo     string
	IssueNum int
}

type Granularity string

const (
	Day  Granularity = "day"
	Week Granularity = "week"
)

type Point struct {
	Date      time.Time
	TokensIn  int
	TokensOut int
	Cost      float64
}

type OverTimeFilter struct {
	Repo        string
	IssueNum    *int
	Agent       string
	Model       string
	From        *time.Time
	To          *time.Time
	Granularity Granularity
}
