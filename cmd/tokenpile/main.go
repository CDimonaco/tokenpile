package main

import (
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/config"
	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/tui"
)

var (
	version            = "dev"
	githubClientID     = ""
	githubClientSecret = ""
)

func main() {
	app := &cli.App{
		Name:    "tokenpile",
		Usage:   "track LLM token usage and cost per GitHub issue",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "log level: debug, info, warn, error",
				Value:   "info",
				EnvVars: []string{"TOKENPILE_LOG_LEVEL"},
			},
			&cli.StringFlag{
				Name:    "log-format",
				Usage:   "log format: text, json",
				Value:   "text",
				EnvVars: []string{"TOKENPILE_LOG_FORMAT"},
			},
		},
		Before: func(c *cli.Context) error {
			initLogging(c.String("log-level"), c.String("log-format"))
			return nil
		},
	}

	paths := config.Resolve()

	if err := config.EnsureDirs(paths); err != nil {
		fmt.Fprintf(os.Stderr, "setup dirs: %v\n", err)
		os.Exit(1)
	}

	priv, _, err := config.EnsureIdentity(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "identity: %v\n", err)
		os.Exit(1)
	}

	pricingLoader, err := pricing.NewLoader(paths.PricingOverride)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pricing: %v\n", err)
		os.Exit(1)
	}

	sqliteStore, err := store.NewSQLiteStore(paths.DBPath, pricingLoader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open store: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if closeErr := sqliteStore.Close(); closeErr != nil {
			slog.Error("close store", "err", closeErr)
		}
	}()

	// env vars override the baked-in values, useful for development
	clientID := githubClientID
	if v := os.Getenv("TOKENPILE_GITHUB_CLIENT_ID"); v != "" {
		clientID = v
	}

	clientSecret := githubClientSecret
	if v := os.Getenv("TOKENPILE_GITHUB_CLIENT_SECRET"); v != "" {
		clientSecret = v
	}

	authProvider := provider.NewGitHubAuthProvider(clientID, clientSecret, paths.CredentialsPath)
	issueProvider := provider.NewGitHubIssueProvider(authProvider)

	app.Commands = []*cli.Command{
		logCommand(sqliteStore),
		reportCommand(sqliteStore),
		authCommands(authProvider),
		pricingCommands(pricingLoader, paths.PricingOverride),
		exportCommands(sqliteStore, priv, version),
		skillCommands(),
	}

	// Override default action to launch TUI, injecting composed deps.
	app.Action = func(c *cli.Context) error {
		token, _ := authProvider.Token(c.Context)

		model := tui.New(sqliteStore, issueProvider, pricingLoader, token)
		p := tea.NewProgram(model, tea.WithAltScreen())

		if _, tuiErr := p.Run(); tuiErr != nil {
			return fmt.Errorf("tui: %w", tuiErr)
		}

		return nil
	}

	if runErr := app.Run(os.Args); runErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", runErr)
		os.Exit(1) //nolint:gocritic
	}
}

func initLogging(level, format string) {
	var lvl slog.Level

	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler

	if format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))
}
