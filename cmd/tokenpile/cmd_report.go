package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/store"
)

func reportCommand(s store.Store) *cli.Command {
	return &cli.Command{
		Name:  "report",
		Usage: "show token usage report for a GitHub issue",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "issue",
				Aliases:  []string{"i"},
				Usage:    "GitHub issue number",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "repo",
				Aliases: []string{"r"},
				Usage:   "repository in owner/repo format (inferred from git remote if absent)",
			},
			&cli.BoolFlag{
				Name:  "sessions",
				Usage: "show per-session breakdown instead of aggregated summary",
			},
		},
		Action: func(c *cli.Context) error {
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
			ctx := c.Context

			if c.Bool("sessions") {
				return printSessionsReport(c, s, repo, issueNum)
			}

			report, err := s.GetReport(ctx, repo, issueNum)
			if err != nil {
				return fmt.Errorf("get report: %w", err)
			}

			cached, cacheErr := s.GetIssueCache(ctx, repo, issueNum)
			if cacheErr != nil && !errors.Is(cacheErr, store.ErrIssueCacheNotFound) {
				return fmt.Errorf("get issue cache: %w", cacheErr)
			}

			fmt.Fprintf(c.App.Writer, "Report: %s #%d\n", repo, issueNum)

			if cached != nil {
				fmt.Fprintf(c.App.Writer, "Title:  %s\n", cached.Title)
				fmt.Fprintf(c.App.Writer, "URL:    https://github.com/%s/issues/%d\n", repo, issueNum)

				if len(cached.Labels) > 0 {
					fmt.Fprintf(c.App.Writer, "Labels: %s\n", strings.Join(cached.Labels, ", "))
				}
			}

			fmt.Fprintln(c.App.Writer)
			fmt.Fprintf(c.App.Writer, "%-16s %-24s %-8s %-12s %-12s %s\n",
				"Agent", "Model", "Calls", "Tokens In", "Tokens Out", "Cost")
			fmt.Fprintf(
				c.App.Writer,
				"%s\n",
				"--------------------------------------------------------------------------------",
			)

			for _, row := range report.Rows {
				fmt.Fprintf(c.App.Writer, "%-16s %-24s %-8d %-12d %-12d $%.6f\n",
					row.Agent, row.Model, row.Calls, row.TokensIn, row.TokensOut, row.Cost)
			}

			fmt.Fprintf(
				c.App.Writer,
				"%s\n",
				"--------------------------------------------------------------------------------",
			)
			fmt.Fprintf(c.App.Writer, "%-41s %-12d %-12d $%.6f\n",
				"Total", report.TotalTokensIn, report.TotalTokensOut, report.TotalCost)
			fmt.Fprintf(c.App.Writer, "\nWall-clock time: %s\n", report.TotalTime.Round(1000000000))

			budget, budgetErr := s.GetBudget(ctx, repo, issueNum)
			if budgetErr != nil && !errors.Is(budgetErr, store.ErrBudgetNotFound) {
				return fmt.Errorf("get budget: %w", budgetErr)
			}

			if budget != nil {
				pct := report.TotalCost / *budget * 100
				fmt.Fprintf(c.App.Writer, "Budget:          $%.2f / $%.2f (%.1f%%)\n",
					report.TotalCost, *budget, pct)
			}

			return nil
		},
	}
}

func printSessionsReport(c *cli.Context, s store.Store, repo string, issueNum int) error {
	ctx := c.Context

	sessions, err := s.ListSessions(ctx, repo, issueNum)
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}

	fmt.Fprintf(c.App.Writer, "Sessions: %s #%d\n\n", repo, issueNum)

	if len(sessions) == 0 {
		fmt.Fprintln(c.App.Writer, "No sessions found.")

		return nil
	}

	sep := strings.Repeat("-", 72)

	for i, sess := range sessions {
		end := "active"
		duration := time.Since(sess.StartedAt).Round(time.Second)

		if sess.EndedAt != nil {
			end = sess.EndedAt.Local().Format("15:04:05")
			duration = sess.EndedAt.Sub(sess.StartedAt).Round(time.Second)
		}

		tags := ""
		if len(sess.Tags) > 0 {
			tags = "[" + strings.Join(sess.Tags, "] [") + "]"
		}

		fmt.Fprintf(c.App.Writer, "Session %d  %s → %s  (%s)  %s\n",
			i+1,
			sess.StartedAt.Local().Format("2006-01-02 15:04:05"),
			end,
			duration,
			tags,
		)

		if sess.Note != "" {
			fmt.Fprintf(c.App.Writer, "  %s\n", sess.Note)
		}

		fmt.Fprintln(c.App.Writer, sep)
	}

	return nil
}
