package main

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/store"
)

func budgetCommands(s store.Store) *cli.Command {
	return &cli.Command{
		Name:  "budget",
		Usage: "manage per-issue spending budgets",
		Subcommands: []*cli.Command{
			{
				Name:  "set",
				Usage: "set a budget (USD) for an issue",
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
					&cli.Float64Flag{
						Name:     "amount",
						Aliases:  []string{"a"},
						Usage:    "budget amount in USD (e.g. 10.00)",
						Required: true,
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

					amount := c.Float64("amount")
					if amount <= 0 {
						return errors.New("--amount must be greater than zero")
					}

					issueNum := c.Int("issue")

					if err = s.SetBudget(c.Context, repo, issueNum, amount); err != nil {
						return fmt.Errorf("set budget: %w", err)
					}

					fmt.Fprintf(c.App.Writer, "Budget set: %s #%d = $%.2f\n", repo, issueNum, amount)

					return nil
				},
			},
			{
				Name:  "unset",
				Usage: "remove the budget for an issue",
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

					if err = s.UnsetBudget(c.Context, repo, issueNum); err != nil {
						return fmt.Errorf("unset budget: %w", err)
					}

					fmt.Fprintf(c.App.Writer, "Budget removed: %s #%d\n", repo, issueNum)

					return nil
				},
			},
		},
	}
}
