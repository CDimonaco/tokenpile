package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/export"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)


func exportCommands(s store.Store, priv ed25519.PrivateKey, version string) *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "export usage data as signed JSON",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "write to file instead of stdout",
			},
			&cli.StringFlag{
				Name:  "repo",
				Usage: "filter by repository (owner/repo)",
			},
			&cli.IntFlag{
				Name:  "issue",
				Usage: "filter by issue number",
			},
			&cli.StringFlag{
				Name:  "agent",
				Usage: "filter by agent name",
			},
			&cli.StringFlag{
				Name:  "model",
				Usage: "filter by model name",
			},
			&cli.StringFlag{
				Name:  "from",
				Usage: "start date filter (RFC3339)",
			},
			&cli.StringFlag{
				Name:  "to",
				Usage: "end date filter (RFC3339)",
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:  "verify",
				Usage: "verify a signed export file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "path to export file",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					path := c.String("file")

					data, err := os.ReadFile(path)
					if err != nil {
						return fmt.Errorf("read file: %w", err)
					}

					var doc export.Document
					if err = json.Unmarshal(data, &doc); err != nil {
						return fmt.Errorf("parse export: %w", err)
					}

					if err = export.Verify(&doc); err != nil {
						fmt.Fprintf(c.App.Writer, "INVALID: %v\n", err)

						return fmt.Errorf("verification failed: %w", err)
					}

					fmt.Fprintf(c.App.Writer, "OK: signature valid, %d entries\n", len(doc.Entries))

					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			return runExport(c, s, priv, version)
		},
	}
}

func runExport(c *cli.Context, s store.Store, priv ed25519.PrivateKey, version string) error {
	ctx := c.Context
	filter := usage.Filter{
		Repo:     strings.ToLower(c.String("repo")),
		IssueNum: c.Int("issue"),
		Agent:    c.String("agent"),
		Model:    c.String("model"),
	}

	if fromStr := c.String("from"); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return fmt.Errorf("parse --from: %w", err)
		}

		filter.From = &t
	}

	if toStr := c.String("to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return fmt.Errorf("parse --to: %w", err)
		}

		filter.To = &t
	}

	entries, err := s.ListEntries(ctx, filter)
	if err != nil {
		return fmt.Errorf("list entries: %w", err)
	}

	doc, err := export.Build(entries, priv, version)
	if err != nil {
		return fmt.Errorf("build export: %w", err)
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal export: %w", err)
	}

	if outPath := c.String("output"); outPath != "" {
		if err = os.WriteFile(outPath, data, 0o600); err != nil {
			return fmt.Errorf("write output: %w", err)
		}

		fmt.Fprintf(c.App.Writer, "Exported %d entries to %s\n", len(entries), outPath)

		return nil
	}

	fmt.Fprintf(c.App.Writer, "%s\n", data)

	return nil
}
