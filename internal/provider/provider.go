package provider

import (
	"context"
	"errors"

	"github.com/cdimonaco/tokenpile/internal/domain"
)

//go:generate mockery --name=AuthProvider --output=../mocks --outpkg=mocks
//go:generate mockery --name=IssueProvider --output=../mocks --outpkg=mocks

var (
	ErrNoRepo          = errors.New("cannot infer repo: not a git repository or no origin remote configured; pass --repo owner/repo")
	ErrUnauthenticated = errors.New("not authenticated: run tokenpile auth login")
)

type AuthProvider interface {
	Login(ctx context.Context) error
	Token(ctx context.Context) (string, error)
	Logout(ctx context.Context) error
}

type IssueProvider interface {
	ListIssues(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error)
	GetIssue(ctx context.Context, repo string, number int) (*domain.Issue, error)
}
