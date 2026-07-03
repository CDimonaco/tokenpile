package main

import (
	"errors"
	"fmt"

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

			report, err := s.GetReport(ctx, repo, issueNum)
			if err != nil {
				return fmt.Errorf("get report: %w", err)
			}

			fmt.Fprintf(c.App.Writer, "Report: %s #%d\n\n", repo, issueNum)
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

			return nil
		},
	}
}
