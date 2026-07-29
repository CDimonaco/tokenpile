package provider

import (
	"context"
	"errors"

	"github.com/cdimonaco/tokenpile/internal/usage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=AuthProvider --output=../mocks --outpkg=mocks --filename=auth_provider_mock.go
//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=IssueProvider --output=../mocks --outpkg=mocks --filename=issue_provider_mock.go

var (
	ErrNoRepo = errors.New(
		"cannot infer repo: not a git repository or no origin remote configured; pass --repo owner/repo",
	)
	ErrUnauthenticated = errors.New("not authenticated: run tokenpile auth login")
	ErrIssueNotFound   = errors.New("issue not found")

	// ErrGhNotFound reports that the gh CLI is not installed. It stays
	// distinct from ErrUnauthenticated because it names a different remedy:
	// telling the user to run tokenpile auth login here would be wrong advice.
	ErrGhNotFound = errors.New(
		"gh CLI not found on PATH: install it from https://cli.github.com, " +
			"or run tokenpile auth login --no-gh-cli to use the OAuth flow",
	)
	// ErrGhUnauthenticated reports that the gh CLI is installed but holds no
	// usable credential. Distinct from ErrUnauthenticated for the same reason.
	ErrGhUnauthenticated = errors.New("gh CLI is not authenticated: run gh auth login")
)

type Issue struct {
	Number int
	Repo   string
	Title  string
	State  string
	URL    string
	Labels []string
}

type AuthProvider interface {
	Login(ctx context.Context) error
	Token(ctx context.Context) (string, error)
	Logout(ctx context.Context) error
}

type IssueProvider interface {
	ListIssues(ctx context.Context, filter usage.Filter) ([]Issue, error)
	GetIssue(ctx context.Context, repo string, number int) (*Issue, error)
}
