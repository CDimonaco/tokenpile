package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/provider"
)

func authCommands(auth provider.AuthProvider) *cli.Command {
	return &cli.Command{
		Name:  "auth",
		Usage: "manage authentication",
		Subcommands: []*cli.Command{
			{
				Name:  "login",
				Usage: "authenticate with a provider",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "provider",
						Aliases:  []string{"p"},
						Usage:    "provider name (e.g. github)",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					if err := auth.Login(c.Context); err != nil {
						return fmt.Errorf("login: %w", err)
					}

					fmt.Fprintln(c.App.Writer, "Authenticated successfully.")

					return nil
				},
			},
			{
				Name:  "logout",
				Usage: "remove stored credentials for a provider",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "provider",
						Aliases:  []string{"p"},
						Usage:    "provider name (e.g. github)",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					if err := auth.Logout(c.Context); err != nil {
						return fmt.Errorf("logout: %w", err)
					}

					fmt.Fprintln(c.App.Writer, "Logged out.")

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "show authentication status",
				Action: func(c *cli.Context) error {
					token, err := auth.Token(c.Context)
					if err != nil || token == "" {
						fmt.Fprintln(c.App.Writer, "github: not logged in")

						return nil
					}

					fmt.Fprintln(c.App.Writer, "github: authenticated")

					return nil
				},
			},
		},
	}
}
