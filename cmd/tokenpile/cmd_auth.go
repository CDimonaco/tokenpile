package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/provider"
)

// ghCredentialSource is what login and status need from the gh CLI: enough to
// detect it, borrow its token and name the account, and no more.
type ghCredentialSource interface {
	Token(ctx context.Context) (string, error)
	GhLogin(ctx context.Context) (string, error)
	Available(ctx context.Context) bool
}

// tokenValidator proves a borrowed token works before tokenpile commits to it.
// Injected so login stays testable without reaching GitHub.
type tokenValidator func(ctx context.Context, token string) (provider.TokenInfo, error)

// authCommands wires both token sources. oauth owns the credential slot, so it
// is what logout clears whichever source is active; gh only ever borrows.
func authCommands(
	oauth provider.AuthProvider,
	gh ghCredentialSource,
	validate tokenValidator,
	credPath string,
) *cli.Command {
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
					&cli.BoolFlag{
						Name:  "use-gh-cli",
						Usage: "use the credential held by the gh CLI instead of the OAuth flow",
					},
					&cli.BoolFlag{
						Name:  "no-gh-cli",
						Usage: "always use the OAuth flow, skipping gh CLI detection",
					},
				},
				Action: func(c *cli.Context) error {
					return runLogin(c, oauth, gh, validate, credPath)
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
					// Clearing the slot covers both an OAuth token and a gh
					// token source marker. gh's own credential is untouched:
					// tokenpile borrowed it, it never owned it.
					if err := oauth.Logout(c.Context); err != nil {
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
					return runAuthStatus(c, oauth, gh, credPath)
				},
			},
		},
	}
}

func runLogin(
	c *cli.Context,
	oauth provider.AuthProvider,
	gh ghCredentialSource,
	validate tokenValidator,
	credPath string,
) error {
	if c.Bool("use-gh-cli") && c.Bool("no-gh-cli") {
		return errors.New("--use-gh-cli and --no-gh-cli are mutually exclusive")
	}

	if shouldUseGhCli(c, gh) {
		return loginWithGhCli(c, gh, validate, credPath)
	}

	if err := oauth.Login(c.Context); err != nil {
		return fmt.Errorf("login: %w", err)
	}

	fmt.Fprintln(c.App.Writer, "Authenticated successfully via the tokenpile OAuth App.")

	return nil
}

// shouldUseGhCli resolves the token source for this login: explicit flags win,
// otherwise an installed and authenticated gh is offered. When gh is absent or
// unusable there is nothing to ask about, so the OAuth flow runs unannounced.
func shouldUseGhCli(c *cli.Context, gh ghCredentialSource) bool {
	if c.Bool("use-gh-cli") {
		return true
	}

	if c.Bool("no-gh-cli") {
		return false
	}

	if !gh.Available(c.Context) {
		return false
	}

	// A gh that cannot name its account has nothing to offer, so there is
	// nothing to ask about: fall through to the OAuth flow unannounced.
	login, ghErr := gh.GhLogin(c.Context)
	if ghErr != nil {
		login = ""
	}

	if login == "" {
		return false
	}

	return promptYesNo(
		c.App.Writer,
		c.App.Reader,
		fmt.Sprintf("Found gh CLI authenticated as %s. Use its credential?", login),
	)
}

func loginWithGhCli(
	c *cli.Context,
	gh ghCredentialSource,
	validate tokenValidator,
	credPath string,
) error {
	token, err := gh.Token(c.Context)
	if err != nil {
		return fmt.Errorf("read gh credential: %w", err)
	}

	// Validate before persisting: a gh credential without repo access fails
	// only at the first issue lookup otherwise, far from the cause.
	info, err := validate(c.Context, token)
	if err != nil {
		return fmt.Errorf("validate gh credential: %w", err)
	}

	if !info.HasScope("repo") {
		return fmt.Errorf(
			"gh credential for %s lacks the repo scope needed to read issues: run gh auth refresh -s repo",
			info.Login,
		)
	}

	if err = provider.PersistGhCliTokenSource(credPath); err != nil {
		return err
	}

	fmt.Fprintf(c.App.Writer, "Using the gh CLI credential, authenticated as %s.\n", info.Login)

	return nil
}

func runAuthStatus(
	c *cli.Context,
	oauth provider.AuthProvider,
	gh ghCredentialSource,
	credPath string,
) error {
	if provider.StoredTokenSource(credPath) == provider.TokenSourceGhCli {
		return ghStatus(c, gh)
	}

	token, _ := oauth.Token(c.Context)
	if token == "" {
		fmt.Fprintln(c.App.Writer, "github: not logged in")

		return nil
	}

	fmt.Fprintln(c.App.Writer, "github: authenticated via the tokenpile OAuth App")

	return nil
}

func ghStatus(c *cli.Context, gh ghCredentialSource) error {
	login, err := gh.GhLogin(c.Context)
	if err != nil {
		// The stored source says gh, but gh cannot answer. Report it rather
		// than quietly falling back: the fallback would hide the breakage.
		fmt.Fprintf(c.App.Writer, "github: token source is the gh CLI, but it is unavailable: %v\n", err)

		return fmt.Errorf("gh CLI token source unavailable: %w", err)
	}

	fmt.Fprintf(c.App.Writer, "github: authenticated via the gh CLI as %s\n", login)

	return nil
}

func promptYesNo(out io.Writer, in io.Reader, question string) bool {
	fmt.Fprintf(out, "%s [Y/n]: ", question)

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && line == "" {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "", "y", "yes":
		return true
	default:
		return false
	}
}
