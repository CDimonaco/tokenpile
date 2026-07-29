package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const ghBinary = "gh"

// GhCliAuthProvider borrows the GitHub credential the gh CLI already holds.
//
// It exists because tokenpile's OAuth App needs org admin approval to reach
// repositories owned by an organization, and that approval is often never
// granted. gh does not bypass those restrictions, but its credential has
// almost always already been approved, so tokenpile can inherit it instead of
// asking for a second approval that will not come.
//
// gh is used strictly as a token source: issue operations still run through
// GitHubIssueProvider and go-github.
type GhCliAuthProvider struct {
	binary string
}

var _ AuthProvider = (*GhCliAuthProvider)(nil)

func NewGhCliAuthProvider() *GhCliAuthProvider {
	return &GhCliAuthProvider{binary: ghBinary}
}

// Login has nothing to authenticate: gh owns the credential. It only reports
// whether that credential is currently usable, so selecting this token source
// fails loudly at login rather than at the first issue lookup.
func (p *GhCliAuthProvider) Login(ctx context.Context) error {
	if _, err := p.Token(ctx); err != nil {
		return err
	}

	return nil
}

// Token shells out on every call rather than caching. The few milliseconds cost
// less than owning an expiry we cannot see: refresh stays with gh.
func (p *GhCliAuthProvider) Token(ctx context.Context) (string, error) {
	stdout, stderr, err := p.run(ctx, "auth", "token")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", withDetail(ErrGhUnauthenticated, stderr)
		}

		return "", err
	}

	if stdout == "" {
		return "", ErrGhUnauthenticated
	}

	return stdout, nil
}

// Logout is a no-op: tokenpile never owns the gh credential, so logging out of
// tokenpile must not log the user out of gh. Clearing the stored token source
// is the auth command's job.
func (p *GhCliAuthProvider) Logout(_ context.Context) error {
	return nil
}

// GhLogin reports the handle gh is authenticated as, for login confirmation and
// auth status output.
func (p *GhCliAuthProvider) GhLogin(ctx context.Context) (string, error) {
	stdout, stderr, err := p.run(ctx, "api", "user", "--jq", ".login")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", withDetail(ErrGhUnauthenticated, stderr)
		}

		return "", err
	}

	return stdout, nil
}

// Available reports whether the gh CLI is installed and holds a usable
// credential, so detection at login can stay silent when it is not.
func (p *GhCliAuthProvider) Available(ctx context.Context) bool {
	_, err := p.Token(ctx)

	return err == nil
}

func (p *GhCliAuthProvider) run(ctx context.Context, args ...string) (string, string, error) {
	if _, err := exec.LookPath(p.binary); err != nil {
		return "", "", ErrGhNotFound
	}

	var stdout, stderr bytes.Buffer

	// #nosec G204 -- p.binary is the ghBinary constant and every args value is
	// a literal from this file; nothing here comes from user input.
	cmd := exec.CommandContext(ctx, p.binary, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outText := strings.TrimSpace(stdout.String())
	errText := strings.TrimSpace(stderr.String())

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return outText, errText, err
		}

		return outText, errText, fmt.Errorf("run gh %s: %w", strings.Join(args, " "), err)
	}

	return outText, errText, nil
}

func withDetail(sentinel error, detail string) error {
	if detail == "" {
		return sentinel
	}

	return fmt.Errorf("%w (gh said: %s)", sentinel, detail)
}
