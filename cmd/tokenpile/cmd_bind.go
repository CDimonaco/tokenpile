package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/attribution"
	"github.com/cdimonaco/tokenpile/internal/config"
	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/store"
)

// bindCommand declares which issue the current work belongs to.
//
// This is the one thing in the pipeline that must come from whoever is doing
// the work: the token counts are measured, but only a person or an agent knows
// which issue a stretch of work was for. Forgetting to bind costs attribution,
// never measurement.
func bindCommand(paths config.Paths) *cli.Command {
	return &cli.Command{
		Name:  "bind",
		Usage: "declare the issue the current session is working on",
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
			&cli.StringFlag{
				Name:  "session",
				Usage: "agent session id (defaults to binding the working directory)",
			},
			&cli.StringFlag{
				Name:  "note",
				Usage: "brief description of what is being worked on",
			},
			&cli.StringSliceFlag{
				Name:  "tag",
				Usage: "label for this work; repeatable",
			},
		},
		Action: func(c *cli.Context) error {
			repo, err := provider.ResolveRepo(c.String("repo"))
			if err != nil {
				return fmt.Errorf("resolve repo: %w", err)
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve working directory: %w", err)
			}

			bindings := attribution.NewStore(paths.BindingsPath)

			if err = bindings.Bind(c.String("session"), cwd, attribution.Binding{
				Repo:     repo,
				IssueNum: c.Int("issue"),
				Note:     c.String("note"),
				Tags:     c.StringSlice("tag"),
			}); err != nil {
				return fmt.Errorf("bind issue: %w", err)
			}

			fmt.Fprintf(c.App.Writer, "Bound %s #%d\n", repo, c.Int("issue"))

			return nil
		},
	}
}

// unattributedCommand lists and assigns usage that belongs to no issue.
//
// Assignment is per session rather than per entry: a session is dozens of
// turns, and assigning them one by one would make reconciliation unusable.
func unattributedCommand(s store.Store, paths config.Paths) *cli.Command {
	return &cli.Command{
		Name:  "unattributed",
		Usage: "list usage recorded without an issue, and assign it",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repo",
				Aliases: []string{"r"},
				Usage:   "limit to a repository",
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:      "assign",
				Usage:     "attribute a session's usage to an issue",
				ArgsUsage: "<session-id>",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "issue",
						Aliases: []string{"i"},
						Usage:   "GitHub issue number (defaults to the branch-derived suggestion)",
					},
				},
				Action: func(c *cli.Context) error {
					sessionID := c.Args().First()
					if sessionID == "" {
						return errors.New("session id is required")
					}

					issueNum, err := resolveAssignIssue(c, s, sessionID)
					if err != nil {
						return err
					}

					n, err := s.AssignIssue(c.Context, sessionID, issueNum)
					if err != nil {
						return fmt.Errorf("assign issue: %w", err)
					}

					fmt.Fprintf(c.App.Writer, "Assigned %d entries to #%d\n", n, issueNum)

					return nil
				},
			},
			{
				Name:      "unassign",
				Usage:     "return a session's usage to unattributed",
				ArgsUsage: "<session-id>",
				Action: func(c *cli.Context) error {
					sessionID := c.Args().First()
					if sessionID == "" {
						return errors.New("session id is required")
					}

					n, err := s.UnassignIssue(c.Context, sessionID)
					if err != nil {
						return fmt.Errorf("unassign issue: %w", err)
					}

					fmt.Fprintf(c.App.Writer, "Returned %d entries to unattributed\n", n)

					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			reconcileSpool(c, s, paths)

			groups, err := s.ListUnattributed(c.Context, c.String("repo"))
			if err != nil {
				return fmt.Errorf("list unattributed: %w", err)
			}

			if len(groups) == 0 {
				fmt.Fprintln(c.App.Writer, "No unattributed usage.")

				return nil
			}

			fmt.Fprintf(c.App.Writer, "%-38s %-22s %-24s %-8s %-12s %-12s %s\n",
				"Session", "Repo", "Branch", "Entries", "Tokens", "Cost", "Suggested")
			fmt.Fprintln(c.App.Writer, strings.Repeat("-", 130))

			for _, g := range groups {
				fmt.Fprintf(c.App.Writer, "%-38s %-22s %-24s %-8d %-12d $%-11.6f %s\n",
					g.SessionID, truncateRepo(g.Repo), truncateBranch(g.Branch),
					g.Entries, g.Usage.TotalTokens(), g.Cost, suggestionLabel(g.Suggested))
			}

			fmt.Fprintln(c.App.Writer,
				"\nAssign with: tokenpile unattributed assign <session-id> --issue <n>"+
					"\nOmit --issue to accept the suggested issue.")

			return nil
		},
	}
}

// resolveAssignIssue takes the issue from the flag, or falls back to the
// suggestion the group's branch implies. The fallback exists so accepting a
// suggestion does not mean retyping a number that was just displayed; it is
// never silent, because the number is printed back on success.
func resolveAssignIssue(c *cli.Context, s store.Store, sessionID string) (int, error) {
	if c.IsSet("issue") {
		if c.Int("issue") <= 0 {
			return 0, errors.New("--issue must be a positive issue number")
		}

		return c.Int("issue"), nil
	}

	groups, err := s.ListUnattributed(c.Context, "")
	if err != nil {
		return 0, fmt.Errorf("list unattributed: %w", err)
	}

	for _, g := range groups {
		if g.SessionID == sessionID && g.Suggested != nil {
			return *g.Suggested, nil
		}
	}

	return 0, errors.New("no issue suggested for this session: pass --issue")
}

func truncateRepo(repo string) string {
	return truncate(repo, 22)
}

func truncateBranch(branch string) string {
	if branch == "" {
		return "-"
	}

	return truncate(branch, 24)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-1] + "…"
}

// suggestionLabel renders the branch-derived suggestion. A dash rather than a
// number is the honest output for a branch that implies nothing: a wrong
// suggestion invites a careless confirmation.
func suggestionLabel(suggested *int) string {
	if suggested == nil {
		return "-"
	}

	return fmt.Sprintf("#%d", *suggested)
}
