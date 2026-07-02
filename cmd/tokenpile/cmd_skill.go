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
						Usage:    "agent name (e.g. claude-code)",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					agentName := c.String("agent")

					path, existed, err := skill.Install(agentName)
					if err != nil {
						return fmt.Errorf("install skill: %w", err)
					}

					if existed {
						fmt.Fprintf(c.App.Writer, "Updated skill for %s at %s\n", agentName, path)
					} else {
						fmt.Fprintf(c.App.Writer, "Installed skill for %s at %s\n", agentName, path)
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
