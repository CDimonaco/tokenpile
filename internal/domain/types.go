package domain

import "time"

type UsageEntry struct {
	ID        string
	Repo      string
	IssueNum  int
	Agent     string
	Model     string
	TokensIn  int
	TokensOut int
	SessionID string
	At        time.Time
}

type Session struct {
	ID        string
	Repo      string
	IssueNum  int
	StartedAt time.Time
	EndedAt   *time.Time
}

type Issue struct {
	Number int
	Repo   string
	Title  string
	State  string
	URL    string
}

type IssueFilter struct {
	Repo     string
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

type TrackedIssue struct {
	IssueNum       int
	Repo           string
	Title          string
	TotalTokensIn  int
	TotalTokensOut int
	TotalCost      float64
	TotalTime      time.Duration
}

type Granularity string

const (
	GranularityDay  Granularity = "day"
	GranularityWeek Granularity = "week"
)

type UsagePoint struct {
	Date      time.Time
	TokensIn  int
	TokensOut int
	Cost      float64
}

type UsageOverTimeFilter struct {
	Repo        string
	IssueNum    *int
	Agent       string
	Model       string
	From        *time.Time
	To          *time.Time
	Granularity Granularity
}
