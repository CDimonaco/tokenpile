package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/skill"
)

func skillCommands() *cli.Command {
	return &cli.Command{
		Name:  "skill",
		Usage: "manage agent skill integrations",
		Subcommands: []*cli.Command{
			{
				Name:  "install",
				Usage: "install skill for an agent",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "agent",
						Aliases:  []string{"a"},
						Usage:    "agent name (claude-code, codex, opencode)",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					agentName := c.String("agent")

					agents := skill.List()
					var agent *skill.Agent
					for i := range agents {
						if agents[i].Name == agentName {
							agent = &agents[i]
							break
						}
					}

					path, existed, err := skill.Install(agentName)
					if err != nil {
						return fmt.Errorf("install skill: %w", err)
					}

					if agent != nil && agent.Shared {
						printSharedInstallResult(c, agentName, path, existed)
					} else {
						printDedicatedInstallResult(c, agentName, path, existed)
					}

					return nil
				},
			},
			{
				Name:  "list",
				Usage: "list supported agents and installation status",
				Action: func(c *cli.Context) error {
					agents := skill.List()

					fmt.Fprintf(c.App.Writer, "%-20s %s\n", "Agent", "Status")
					fmt.Fprintf(c.App.Writer, "%s\n", "------------------------------")

					for _, a := range agents {
						status := "not installed"
						if skill.IsInstalled(a.Name) {
							status = "installed"
						}

						fmt.Fprintf(c.App.Writer, "%-20s %s\n", a.Name, status)
					}

					return nil
				},
			},
		},
	}
}

func printDedicatedInstallResult(c *cli.Context, agentName, path string, existed bool) {
	if existed {
		fmt.Fprintf(c.App.Writer, "Updated skill for %s at %s\n", agentName, path)
	} else {
		fmt.Fprintf(c.App.Writer, "Installed skill for %s at %s\n", agentName, path)
	}
}

func printSharedInstallResult(c *cli.Context, agentName, path string, existed bool) {
	fmt.Fprintf(c.App.Writer, "Installing tokenpile skill for %s\n", agentName)
	fmt.Fprintf(c.App.Writer, "  Target file: %s\n\n", path)

	if existed {
		fmt.Fprintf(c.App.Writer, "  The file already contained a tokenpile block.\n")
		fmt.Fprintf(c.App.Writer, "  It has been updated in place — the rest of your %s is unchanged.\n\n", "AGENTS.md")
	} else {
		fmt.Fprintf(c.App.Writer, "  Appended a tokenpile block to %s.\n", "AGENTS.md")
		fmt.Fprintf(c.App.Writer, "  Your existing instructions were not modified.\n\n")
	}

	fmt.Fprintf(c.App.Writer, "  The block is delimited by:\n")
	fmt.Fprintf(c.App.Writer, "    <!-- tokenpile:start --> ... <!-- tokenpile:end -->\n")
	fmt.Fprintf(c.App.Writer, "  Running this command again will update only that block.\n\n")
	fmt.Fprintf(c.App.Writer, "Skill ready. %s will log usage automatically.\n", agentName)
}
