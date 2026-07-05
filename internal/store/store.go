package store

import (
	"context"
	"time"

	"github.com/cdimonaco/tokenpile/internal/usage"
)

//go:generate mockery --name=Store --output=../mocks --outpkg=mocks

type Store interface {
	LogUsage(ctx context.Context, entry usage.Entry) error
	ListEntries(ctx context.Context, filter usage.Filter) ([]usage.Entry, error)
	ListIssues(ctx context.Context, filter usage.Filter) ([]usage.TrackedIssue, error)
	GetReport(ctx context.Context, repo string, issueNum int) (*usage.Report, error)
	UpsertIssueCache(ctx context.Context, cache *usage.IssueCache) error
	GetIssueCache(ctx context.Context, repo string, issueNum int) (*usage.IssueCache, error)
	StartSession(ctx context.Context, repo string, issueNum int) (*usage.Session, error)
	EndSession(ctx context.Context, sessionID string) error
	EndSessionAt(ctx context.Context, sessionID string, at time.Time) error
	ListSessions(ctx context.Context, repo string, issueNum int) ([]usage.Session, error)
	UpdateSessionAnnotations(ctx context.Context, sessionID string, note *string, tags []string) error
	UpdateSessionActivity(ctx context.Context, sessionID string, at time.Time) error
	ListUsageOverTime(ctx context.Context, filter usage.OverTimeFilter) ([]usage.Point, error)
	ListTrackedIssueRefs(ctx context.Context) ([]usage.TrackedIssueRef, error)
	SetBudget(ctx context.Context, repo string, issueNum int, budget float64) error
	UnsetBudget(ctx context.Context, repo string, issueNum int) error
	GetBudget(ctx context.Context, repo string, issueNum int) (*float64, error)
}
