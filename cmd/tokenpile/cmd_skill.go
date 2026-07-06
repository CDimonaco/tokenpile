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

					path, existed, err := skill.Install(agentName)
					if err != nil {
						return fmt.Errorf("install skill: %w", err)
					}

					printDedicatedInstallResult(c, agentName, path, existed)

					return nil
				},
			},
			{
				Name:  "list",
				Usage: "list supported agents and installation status",
				Action: func(c *cli.Context) error {
					agents := skill.List()

					fmt.Fprintf(c.App.Writer, "%-20s %-16s %s\n", "Agent", "Status", "Skill version")
					fmt.Fprintf(c.App.Writer, "%s\n", "------------------------------------------------")

					for _, a := range agents {
						installed := skill.IsInstalled(a.Name)
						status := "not installed"
						versionNote := ""

						if installed {
							status = "installed"
							if !skill.IsUpToDate(a.Name) {
								versionNote = "outdated — run: tokenpile skill install --agent " + a.Name
							} else {
								versionNote = "up to date"
							}
						}

						fmt.Fprintf(c.App.Writer, "%-20s %-16s %s\n", a.Name, status, versionNote)
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
