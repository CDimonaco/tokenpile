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
				Name:  "input",
				Usage: "fresh input tokens, billed at the full input rate",
			},
			&cli.IntFlag{
				Name:  "cache-write",
				Usage: "tokens written to the prompt cache",
			},
			&cli.IntFlag{
				Name:  "cache-read",
				Usage: "tokens served from the prompt cache",
			},
			&cli.IntFlag{
				Name:  "output",
				Usage: "output tokens",
			},
			&cli.IntFlag{
				Name:  "reasoning",
				Usage: "reasoning tokens, a subset of output (not billed on top of it)",
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
	ctx := c.Context

	u := usage.Usage{
		InputFresh: c.Int("input"),
		CacheWrite: c.Int("cache-write"),
		CacheRead:  c.Int("cache-read"),
		Output:     c.Int("output"),
		Reasoning:  c.Int("reasoning"),
	}

	for name, v := range map[string]int{
		"--input": u.InputFresh, "--cache-write": u.CacheWrite,
		"--cache-read": u.CacheRead, "--output": u.Output, "--reasoning": u.Reasoning,
	} {
		if v < 0 {
			return fmt.Errorf("%s must be zero or greater", name)
		}
	}

	if u.TotalTokens() == 0 {
		return errors.New(
			"at least one token count is required: --input, --cache-write, --cache-read or --output",
		)
	}

	if u.Reasoning > u.Output {
		return errors.New("--reasoning cannot exceed --output: reasoning tokens are a subset of output")
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
		ID:       uuid.NewString(),
		Repo:     repo,
		IssueNum: issueNum,
		Agent:    agent,
		Model:    model,
		Usage:    u,
		// log is the manual path: a model cannot observe its own cache tiers,
		// so anything arriving here is a declaration, not a measurement.
		Source:    usage.SourceEstimated,
		SessionID: sessionID,
		At:        time.Now().UTC(),
	}

	if err = s.LogUsage(ctx, entry); err != nil {
		return fmt.Errorf("log usage: %w", err)
	}

	applyAnnotations(ctx, s, sessionID, c.String("note"), c.StringSlice("tag"))

	fmt.Fprintf(c.App.Writer, "Logged: %s #%d  in=%d out=%d  session=%s\n",
		repo, issueNum, u.TotalInput(), u.Output, sessionID)

	return nil
}

func applyAnnotations(ctx context.Context, s store.Store, sessionID, noteStr string, tags []string) {
	if noteStr == "" && len(tags) == 0 {
		return
	}

	if runes := []rune(noteStr); len(runes) > 200 {
		noteStr = string(runes[:200])
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
