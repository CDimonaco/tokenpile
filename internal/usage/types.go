package usage

import "time"

// Usage is the token accounting for a single entry, split by the tier the
// provider bills each count under. Providers price these very differently:
// cached reads cost a fraction of fresh input, cache writes cost more than it.
//
// Reasoning is a SUBSET of Output, never a sixth bucket added on top of it.
// Providers that report reasoning tokens report them inside the completion
// count, so adding the two double-bills every reasoning token.
type Usage struct {
	InputFresh int `json:"input_fresh"`
	CacheWrite int `json:"cache_write"`
	CacheRead  int `json:"cache_read"`
	Output     int `json:"output"`
	Reasoning  int `json:"reasoning"`
}

// TotalTokens is every token that passed through, counted once. Reasoning is
// excluded because it is already inside Output.
func (u Usage) TotalTokens() int {
	return u.InputFresh + u.CacheWrite + u.CacheRead + u.Output
}

// TotalInput is the input-side total across all three input tiers, for displays
// that want a single "in" figure alongside Output.
func (u Usage) TotalInput() int {
	return u.InputFresh + u.CacheWrite + u.CacheRead
}

// Add accumulates another usage into this one, tier by tier.
func (u Usage) Add(other Usage) Usage {
	return Usage{
		InputFresh: u.InputFresh + other.InputFresh,
		CacheWrite: u.CacheWrite + other.CacheWrite,
		CacheRead:  u.CacheRead + other.CacheRead,
		Output:     u.Output + other.Output,
		Reasoning:  u.Reasoning + other.Reasoning,
	}
}

// Source records how an entry's token counts were obtained. It is derived from
// the command that writes the entry, never from a flag: the distinction is
// structural, since an agent cannot observe its own cache tiers.
type Source string

const (
	// SourceMeasured means the counts came from an agent's own transcript, as
	// reported by the provider.
	SourceMeasured Source = "measured"
	// SourceEstimated means the counts were declared by a model.
	SourceEstimated Source = "estimated"
)

type Entry struct {
	ID   string
	Repo string
	// IssueNum is nil when the entry is not attributed to an issue. Capture
	// must be able to record a turn before the issue is known: an unattributed
	// measurement can be assigned later, a discarded one is gone.
	IssueNum    *int
	Agent       string
	Model       string
	Usage       Usage
	Source      Source
	SessionID   string
	At          time.Time
	IssueTitle  string
	IssueLabels []string
}

type Session struct {
	ID   string
	Repo string
	// IssueNum is nil until the session is attributed to an issue.
	IssueNum       *int
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
	Agent string
	Model string
	Calls int
	Usage Usage
	Cost  float64
	// CostIncomplete is true when a tier present in Usage had no rate for this
	// model, so Cost omits it. An incomplete figure that says so beats a
	// confident wrong one.
	CostIncomplete bool
}

type Report struct {
	IssueNum   int
	Repo       string
	Rows       []ReportRow
	TotalUsage Usage
	TotalCost  float64
	// CacheSavings is what the cached tokens would have cost at the fresh input
	// rate, minus what they actually cost.
	CacheSavings float64
	TotalTime    time.Duration
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
	IssueNum   int
	Repo       string
	Title      string
	Labels     []string
	TotalUsage Usage
	TotalCost  float64
	TotalTime  time.Duration
	Budget     *float64
}

type TrackedIssueRef struct {
	Repo     string
	IssueNum int
}

// IssueBudget is a stored budget row for an issue.
type IssueBudget struct {
	Repo     string
	IssueNum int
	Amount   float64
}

type Granularity string

const (
	Day  Granularity = "day"
	Week Granularity = "week"
)

type Point struct {
	Date  time.Time
	Usage Usage
	Cost  float64
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

// UnattributedGroup is a batch of usage recorded without an issue, grouped so
// it can be assigned in one action. A session is many turns, and assigning
// entries one at a time would make reconciliation unusable.
type UnattributedGroup struct {
	Repo      string
	Branch    string
	SessionID string
	Entries   int
	Usage     Usage
	Cost      float64
	First     time.Time
	Last      time.Time
	// Suggested is the issue the branch name implies, if any.
	Suggested *int
}
