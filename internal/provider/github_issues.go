package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"

	"github.com/cdimonaco/tokenpile/internal/usage"
)

type GitHubIssueProvider struct {
	auth    AuthProvider
	baseURL string
}

func NewGitHubIssueProvider(auth AuthProvider) *GitHubIssueProvider {
	return &GitHubIssueProvider{auth: auth}
}

func NewGitHubIssueProviderWithURL(auth AuthProvider, baseURL string) *GitHubIssueProvider {
	return &GitHubIssueProvider{auth: auth, baseURL: baseURL}
}

func (p *GitHubIssueProvider) client(ctx context.Context) (*github.Client, error) {
	token, err := p.auth.Token(ctx)
	if err != nil {
		return nil, ErrUnauthenticated
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	if p.baseURL != "" {
		var parseErr error

		client, parseErr = client.WithAuthToken(token).WithEnterpriseURLs(p.baseURL+"/", p.baseURL+"/")
		if parseErr != nil {
			return nil, fmt.Errorf("set base URL: %w", parseErr)
		}
	}

	return client, nil
}

func (p *GitHubIssueProvider) ListIssues(ctx context.Context, filter usage.Filter) ([]Issue, error) {
	client, err := p.client(ctx)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(filter.Repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format %q: expected owner/repo", filter.Repo)
	}

	owner, repo := parts[0], parts[1]

	opts := &github.IssueListByRepoOptions{
		State:    filter.State,
		Assignee: filter.Assignee,
	}

	if opts.State == "" {
		opts.State = "open"
	}

	ghIssues, _, err := client.Issues.ListByRepo(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}

	issues := make([]Issue, 0, len(ghIssues))

	for _, gi := range ghIssues {
		if gi.PullRequestLinks != nil {
			continue
		}

		labels := make([]string, 0, len(gi.Labels))
		for _, l := range gi.Labels {
			labels = append(labels, l.GetName())
		}

		issues = append(issues, Issue{
			Number: gi.GetNumber(),
			Repo:   filter.Repo,
			Title:  gi.GetTitle(),
			State:  gi.GetState(),
			URL:    gi.GetHTMLURL(),
			Labels: labels,
		})
	}

	return issues, nil
}

func (p *GitHubIssueProvider) GetIssue(ctx context.Context, repo string, number int) (*Issue, error) {
	client, err := p.client(ctx)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format %q: expected owner/repo", repo)
	}

	owner, repoName := parts[0], parts[1]

	gi, _, err := client.Issues.Get(ctx, owner, repoName, number)
	if err != nil {
		var ghErr *github.ErrorResponse
		if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
			return nil, ErrIssueNotFound
		}

		return nil, fmt.Errorf("get issue: %w", err)
	}

	labels := make([]string, 0, len(gi.Labels))
	for _, l := range gi.Labels {
		labels = append(labels, l.GetName())
	}

	return &Issue{
		Number: gi.GetNumber(),
		Repo:   repo,
		Title:  gi.GetTitle(),
		State:  gi.GetState(),
		URL:    gi.GetHTMLURL(),
		Labels: labels,
	}, nil
}
