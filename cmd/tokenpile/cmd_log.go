package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

const sessionIdleTimeout = 30 * time.Minute

func logCommand(s store.Store, ip provider.IssueProvider) *cli.Command {
	return &cli.Command{
		Name:  "log",
		Usage: "record LLM token usage for a GitHub issue",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "issue",
				Aliases:  []string{"i"},
				Usage:    "GitHub issue number",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "agent",
				Aliases:  []string{"a"},
				Usage:    "agent name (e.g. claude-code, opencode)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "model",
				Aliases:  []string{"m"},
				Usage:    "model identifier (e.g. claude-sonnet-4-6)",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "tokens-in",
				Usage:    "number of input tokens",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "tokens-out",
				Usage:    "number of output tokens",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "repo",
				Aliases: []string{"r"},
				Usage:   "repository in owner/repo format (inferred from git remote if absent)",
			},
			&cli.StringFlag{
				Name:  "note",
				Usage: "brief description of what was done (max 200 chars)",
			},
			&cli.StringSliceFlag{
				Name:  "tag",
				Usage: "categorical tag (repeatable): refactor, debug, feature, test, docs, spike, review",
			},
		},
		Action: func(c *cli.Context) error {
			return runLog(c, s, ip)
		},
	}
}

func runLog(c *cli.Context, s store.Store, ip provider.IssueProvider) error {
	repo, err := provider.ResolveRepo(c.String("repo"))
	if err != nil {
		if errors.Is(err, provider.ErrNoRepo) {
			return errors.New(
				"cannot infer repo: pass --repo owner/repo or run from inside a GitHub repository",
			)
		}

		return fmt.Errorf("infer repo: %w", err)
	}

	issueNum := c.Int("issue")
	agent := c.String("agent")
	model := c.String("model")
	tokensIn := c.Int("tokens-in")
	tokensOut := c.Int("tokens-out")
	ctx := c.Context

	if tokensIn < 0 {
		return errors.New("--tokens-in must be zero or greater")
	}

	if tokensOut < 0 {
		return errors.New("--tokens-out must be zero or greater")
	}

	issue, getErr := ip.GetIssue(ctx, repo, issueNum)
	if getErr != nil {
		if errors.Is(getErr, provider.ErrIssueNotFound) {
			return fmt.Errorf("issue #%d not found in %s", issueNum, repo)
		}

		if errors.Is(getErr, provider.ErrUnauthenticated) {
			return errors.New("GitHub authentication required to validate issues: run tokenpile auth login")
		}

		return fmt.Errorf("validate issue: %w", getErr)
	}

	if cacheErr := s.UpsertIssueCache(ctx, &usage.IssueCache{
		Repo:     repo,
		IssueNum: issueNum,
		Title:    issue.Title,
		Labels:   issue.Labels,
	}); cacheErr != nil {
		slog.Warn("upsert issue cache", "err", cacheErr)
	}

	sessionID, err := resolveSession(ctx, s, repo, issueNum)
	if err != nil {
		return fmt.Errorf("resolve session: %w", err)
	}

	entry := usage.Entry{
		ID:        uuid.NewString(),
		Repo:      repo,
		IssueNum:  issueNum,
		Agent:     agent,
		Model:     model,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
		SessionID: sessionID,
		At:        time.Now().UTC(),
	}

	if err = s.LogUsage(ctx, entry); err != nil {
		return fmt.Errorf("log usage: %w", err)
	}

	applyAnnotations(ctx, s, sessionID, c.String("note"), c.StringSlice("tag"))

	fmt.Fprintf(c.App.Writer, "Logged: %s #%d  in=%d out=%d  session=%s\n",
		repo, issueNum, tokensIn, tokensOut, sessionID)

	return nil
}

func applyAnnotations(ctx context.Context, s store.Store, sessionID, noteStr string, tags []string) {
	if noteStr == "" && len(tags) == 0 {
		return
	}

	if len(noteStr) > 200 {
		noteStr = noteStr[:200]
	}

	var notePtr *string
	if noteStr != "" {
		notePtr = &noteStr
	}

	if err := s.UpdateSessionAnnotations(ctx, sessionID, notePtr, tags); err != nil {
		slog.Warn("update session annotations", "err", err)
	}
}

func resolveSession(ctx context.Context, s store.Store, repo string, issueNum int) (string, error) {
	sessions, err := s.ListSessions(ctx, repo, issueNum)
	if err != nil {
		return "", fmt.Errorf("list sessions: %w", err)
	}

	now := time.Now()
	idleThreshold := now.Add(-sessionIdleTimeout)
	activeID := ""

	for _, sess := range sessions {
		if sess.EndedAt != nil {
			continue
		}

		if sess.LastActivityAt.Before(idleThreshold) {
			if endErr := s.EndSessionAt(ctx, sess.ID, sess.LastActivityAt); endErr != nil {
				return "", fmt.Errorf("end idle session: %w", endErr)
			}
		} else {
			activeID = sess.ID
		}
	}

	if activeID != "" {
		if actErr := s.UpdateSessionActivity(ctx, activeID, now); actErr != nil {
			slog.Warn("update session activity", "err", actErr)
		}

		return activeID, nil
	}

	newSess, err := s.StartSession(ctx, repo, issueNum)
	if err != nil {
		return "", fmt.Errorf("start session: %w", err)
	}

	if actErr := s.UpdateSessionActivity(ctx, newSess.ID, now); actErr != nil {
		slog.Warn("update session activity", "err", actErr)
	}

	return newSess.ID, nil
}
