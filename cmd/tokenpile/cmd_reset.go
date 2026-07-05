package main

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/config"
	"github.com/cdimonaco/tokenpile/internal/export"
	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/skill"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

func resetCommand(
	s store.Store,
	paths config.Paths,
	auth provider.AuthProvider,
	priv ed25519.PrivateKey,
	version string,
) *cli.Command {
	return &cli.Command{
		Name:  "reset",
		Usage: "back up all data to a signed export, then delete all local tokenpile state",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "yes",
				Usage: "skip the interactive confirmation",
			},
			&cli.BoolFlag{
				Name:  "no-backup",
				Usage: "skip the signed backup before deleting",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "backup file path (default tokenpile-backup-<timestamp>.json)",
			},
		},
		Action: func(c *cli.Context) error {
			return runReset(c, s, paths, auth, priv, version)
		},
	}
}

func runReset(
	c *cli.Context,
	s store.Store,
	paths config.Paths,
	auth provider.AuthProvider,
	priv ed25519.PrivateKey,
	version string,
) error {
	ctx := c.Context

	items := enumerateResetItems(ctx, paths, auth)
	if len(items) == 0 {
		fmt.Fprintln(c.App.Writer, "Nothing to reset: no local tokenpile state found.")
	} else {
		fmt.Fprintln(c.App.Writer, "The following will be deleted:")

		for _, item := range items {
			fmt.Fprintf(c.App.Writer, "  - %s\n", item)
		}
	}

	if !c.Bool("yes") && !confirmReset(c) {
		return errors.New("reset aborted")
	}

	if !c.Bool("no-backup") {
		backupPath, err := writeResetBackup(ctx, c, s, priv, version)
		if err != nil {
			return fmt.Errorf("backup failed, nothing was deleted: %w", err)
		}

		fmt.Fprintf(c.App.Writer, "Backup written to %s\n", backupPath)
	}

	failures := destroyState(ctx, c, paths, auth)
	if len(failures) > 0 {
		for _, f := range failures {
			fmt.Fprintf(c.App.Writer, "FAILED: %v\n", f)
		}

		return fmt.Errorf("reset incomplete: %d item(s) could not be removed", len(failures))
	}

	fmt.Fprintln(c.App.Writer, "Reset complete.")

	return nil
}

func enumerateResetItems(ctx context.Context, paths config.Paths, auth provider.AuthProvider) []string {
	var items []string

	for _, path := range resetFilePaths(paths) {
		if _, err := os.Stat(path); err == nil {
			items = append(items, path)
		}
	}

	if _, err := auth.Token(ctx); err == nil {
		items = append(items, "GitHub token (OS keychain / credentials file)")
	}

	for _, agent := range skill.List() {
		if skill.IsInstalled(agent.Name) {
			items = append(items, fmt.Sprintf("agent skill: %s (%s)", agent.Name, agent.InstallPath()))
		}
	}

	return items
}

// resetFilePaths lists every file reset removes, including SQLite sidecars.
func resetFilePaths(paths config.Paths) []string {
	return []string{
		paths.DBPath,
		paths.DBPath + "-wal",
		paths.DBPath + "-shm",
		paths.IdentityKeyPath,
		paths.IdentityPubPath,
		paths.CredentialsPath,
		paths.PricingOverride,
	}
}

func confirmReset(c *cli.Context) bool {
	fmt.Fprint(c.App.Writer, "Type 'yes' to confirm: ")

	line, err := bufio.NewReader(c.App.Reader).ReadString('\n')
	if err != nil && line == "" {
		return false
	}

	return strings.TrimSpace(line) == "yes"
}

func writeResetBackup(
	ctx context.Context,
	c *cli.Context,
	s store.Store,
	priv ed25519.PrivateKey,
	version string,
) (string, error) {
	entries, err := s.ListEntries(ctx, usage.Filter{})
	if err != nil {
		return "", fmt.Errorf("list entries: %w", err)
	}

	sessions, err := s.ListAllSessions(ctx)
	if err != nil {
		return "", fmt.Errorf("list sessions: %w", err)
	}

	rows, err := s.ListBudgets(ctx)
	if err != nil {
		return "", fmt.Errorf("list budgets: %w", err)
	}

	budgets := make([]export.IssueBudget, 0, len(rows))
	for _, b := range rows {
		budgets = append(budgets, export.IssueBudget(b))
	}

	doc, err := export.Build(entries, sessions, budgets, priv, version)
	if err != nil {
		return "", fmt.Errorf("build export: %w", err)
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal export: %w", err)
	}

	path := c.String("output")
	if path == "" {
		path = fmt.Sprintf("tokenpile-backup-%s.json", time.Now().UTC().Format("20060102-150405"))
	}

	if err = os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("write backup: %w", err)
	}

	return path, nil
}

func destroyState(
	ctx context.Context,
	c *cli.Context,
	paths config.Paths,
	auth provider.AuthProvider,
) []error {
	var failures []error

	if err := auth.Logout(ctx); err != nil {
		failures = append(failures, fmt.Errorf("logout: %w", err))
	} else {
		fmt.Fprintln(c.App.Writer, "Removed: GitHub token and credentials")
	}

	for _, agent := range skill.List() {
		path, removed, err := skill.Uninstall(agent.Name)
		if err != nil {
			failures = append(failures, fmt.Errorf("uninstall skill %s: %w", agent.Name, err))
			continue
		}

		if removed {
			fmt.Fprintf(c.App.Writer, "Removed: agent skill %s (%s)\n", agent.Name, path)
		}
	}

	for _, path := range resetFilePaths(paths) {
		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			failures = append(failures, fmt.Errorf("remove %s: %w", path, err))

			continue
		}

		fmt.Fprintf(c.App.Writer, "Removed: %s\n", path)
	}

	return failures
}
