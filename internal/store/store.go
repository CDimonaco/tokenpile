package store

import (
	"context"

	"github.com/cdimonaco/tokenpile/internal/domain"
)

//go:generate mockery --name=Store --output=../mocks --outpkg=mocks

type Store interface {
	LogUsage(ctx context.Context, entry domain.UsageEntry) error
	ListIssues(ctx context.Context, filter domain.IssueFilter) ([]domain.TrackedIssue, error)
	GetReport(ctx context.Context, repo string, issueNum int) (*domain.Report, error)
	StartSession(ctx context.Context, repo string, issueNum int) (*domain.Session, error)
	EndSession(ctx context.Context, sessionID string) error
	ListSessions(ctx context.Context, repo string, issueNum int) ([]domain.Session, error)
	ListUsageOverTime(ctx context.Context, filter domain.UsageOverTimeFilter) ([]domain.UsagePoint, error)
}
