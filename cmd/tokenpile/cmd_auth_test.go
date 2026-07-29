package main

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"github.com/zalando/go-keyring"

	"github.com/cdimonaco/tokenpile/internal/mocks"
	"github.com/cdimonaco/tokenpile/internal/provider"
)

// stubGh stands in for the gh CLI without executing anything.
type stubGh struct {
	token     string
	login     string
	err       error
	available bool
}

func (s stubGh) Token(context.Context) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	return s.token, nil
}

func (s stubGh) GhLogin(context.Context) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	return s.login, nil
}

func (s stubGh) Available(context.Context) bool { return s.available }

func okValidator(scopes []string) tokenValidator {
	return func(context.Context, string) (provider.TokenInfo, error) {
		return provider.TokenInfo{Login: "octocat", Scopes: scopes, ScopesKnown: scopes != nil}, nil
	}
}

type authFixture struct {
	oauth    *mocks.AuthProvider
	gh       ghCredentialSource
	validate tokenValidator
	credPath string
	stdin    string
}

func newAuthFixture(t *testing.T) *authFixture {
	t.Helper()
	keyring.MockInit()

	return &authFixture{
		oauth:    &mocks.AuthProvider{},
		gh:       stubGh{},
		validate: okValidator([]string{"repo"}),
		credPath: filepath.Join(t.TempDir(), "credentials"),
	}
}

func (f *authFixture) run(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Reader:   strings.NewReader(f.stdin),
		Commands: []*cli.Command{authCommands(f.oauth, f.gh, f.validate, f.credPath)},
	}

	err := app.RunContext(context.Background(), append([]string{"tok"}, args...))

	return buf.String(), err
}

func TestAuthLogin_OAuthSuccess(t *testing.T) {
	f := newAuthFixture(t)
	f.oauth.On("Login", context.Background()).Return(nil)

	out, err := f.run(t, "auth", "login", "--provider", "github")

	require.NoError(t, err)
	assert.Contains(t, out, "Authenticated")
	f.oauth.AssertExpectations(t)
}

func TestAuthLogin_OAuthFailure(t *testing.T) {
	f := newAuthFixture(t)
	f.oauth.On("Login", context.Background()).Return(errors.New("oauth failed"))

	_, err := f.run(t, "auth", "login", "--provider", "github")

	require.Error(t, err)
	f.oauth.AssertExpectations(t)
}

func TestAuthLogin_MutuallyExclusiveFlags(t *testing.T) {
	f := newAuthFixture(t)

	_, err := f.run(t, "auth", "login", "--provider", "github", "--use-gh-cli", "--no-gh-cli")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestAuthLogin_UseGhCliPersistsSentinel(t *testing.T) {
	f := newAuthFixture(t)
	f.gh = stubGh{token: "gho_gh", login: "octocat", available: true}

	out, err := f.run(t, "auth", "login", "--provider", "github", "--use-gh-cli")

	require.NoError(t, err)
	assert.Contains(t, out, "octocat")
	assert.Equal(t, provider.TokenSourceGhCli, provider.StoredTokenSource(f.credPath))
	// No browser flow was started.
	f.oauth.AssertNotCalled(t, "Login", context.Background())
}

func TestAuthLogin_NoGhCliSkipsDetection(t *testing.T) {
	f := newAuthFixture(t)
	f.gh = stubGh{token: "gho_gh", login: "octocat", available: true}
	f.oauth.On("Login", context.Background()).Return(nil)

	_, err := f.run(t, "auth", "login", "--provider", "github", "--no-gh-cli")

	require.NoError(t, err)
	assert.Equal(t, provider.TokenSourceOAuth, provider.StoredTokenSource(f.credPath))
	f.oauth.AssertExpectations(t)
}

func TestAuthLogin_GhAbsentRunsOAuthWithoutPrompt(t *testing.T) {
	f := newAuthFixture(t)
	f.gh = stubGh{available: false}
	f.oauth.On("Login", context.Background()).Return(nil)

	out, err := f.run(t, "auth", "login", "--provider", "github")

	require.NoError(t, err)
	assert.NotContains(t, out, "Use its credential?")
	f.oauth.AssertExpectations(t)
}

func TestAuthLogin_DetectionPrompt(t *testing.T) {
	tests := []struct {
		name       string
		stdin      string
		wantSource provider.TokenSource
		wantOAuth  bool
	}{
		{name: "empty input accepts the default", stdin: "\n", wantSource: provider.TokenSourceGhCli},
		{name: "explicit yes", stdin: "y\n", wantSource: provider.TokenSourceGhCli},
		{name: "declining falls to oauth", stdin: "n\n", wantSource: provider.TokenSourceOAuth, wantOAuth: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newAuthFixture(t)
			f.gh = stubGh{token: "gho_gh", login: "octocat", available: true}
			f.stdin = tc.stdin

			if tc.wantOAuth {
				f.oauth.On("Login", context.Background()).Return(nil)
			}

			out, err := f.run(t, "auth", "login", "--provider", "github")

			require.NoError(t, err)
			assert.Contains(t, out, "octocat")
			assert.Equal(t, tc.wantSource, provider.StoredTokenSource(f.credPath))
			f.oauth.AssertExpectations(t)
		})
	}
}

func TestAuthLogin_MissingRepoScopeRefused(t *testing.T) {
	f := newAuthFixture(t)
	f.gh = stubGh{token: "gho_gh", login: "octocat", available: true}
	f.validate = okValidator([]string{"read:user"})

	_, err := f.run(t, "auth", "login", "--provider", "github", "--use-gh-cli")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gh auth refresh -s repo")
	assert.Equal(t, provider.TokenSourceOAuth, provider.StoredTokenSource(f.credPath),
		"a refused credential must not change the stored token source")
}

// Fine-grained PATs report no scopes at all; that silence must not be read as
// a missing scope, since those are exactly the credentials users fall back to.
func TestAuthLogin_UnknownScopesAccepted(t *testing.T) {
	f := newAuthFixture(t)
	f.gh = stubGh{token: "github_pat_x", login: "octocat", available: true}
	f.validate = okValidator(nil)

	_, err := f.run(t, "auth", "login", "--provider", "github", "--use-gh-cli")

	require.NoError(t, err)
	assert.Equal(t, provider.TokenSourceGhCli, provider.StoredTokenSource(f.credPath))
}

func TestAuthLogin_ValidationFailureDoesNotPersist(t *testing.T) {
	f := newAuthFixture(t)
	f.gh = stubGh{token: "gho_gh", login: "octocat", available: true}
	f.validate = func(context.Context, string) (provider.TokenInfo, error) {
		return provider.TokenInfo{}, errors.New("401 bad credentials")
	}

	_, err := f.run(t, "auth", "login", "--provider", "github", "--use-gh-cli")

	require.Error(t, err)
	assert.Equal(t, provider.TokenSourceOAuth, provider.StoredTokenSource(f.credPath))
}

func TestAuthStatus_OAuth(t *testing.T) {
	f := newAuthFixture(t)
	f.oauth.On("Token", context.Background()).Return("tok123", nil)

	out, err := f.run(t, "auth", "status")

	require.NoError(t, err)
	assert.Contains(t, out, "OAuth App")
	f.oauth.AssertExpectations(t)
}

func TestAuthStatus_NotLoggedIn(t *testing.T) {
	f := newAuthFixture(t)
	f.oauth.On("Token", context.Background()).Return("", errors.New("not found"))

	out, err := f.run(t, "auth", "status")

	require.NoError(t, err)
	assert.Contains(t, out, "not logged in")
	f.oauth.AssertExpectations(t)
}

func TestAuthStatus_GhCli(t *testing.T) {
	f := newAuthFixture(t)
	f.gh = stubGh{login: "octocat", available: true}
	require.NoError(t, provider.PersistGhCliTokenSource(f.credPath))

	out, err := f.run(t, "auth", "status")

	require.NoError(t, err)
	assert.Contains(t, out, "gh CLI")
	assert.Contains(t, out, "octocat")
}

func TestAuthStatus_GhCliBrokenExitsNonZero(t *testing.T) {
	f := newAuthFixture(t)
	f.gh = stubGh{err: provider.ErrGhNotFound}
	require.NoError(t, provider.PersistGhCliTokenSource(f.credPath))

	out, err := f.run(t, "auth", "status")

	require.Error(t, err)
	assert.Contains(t, out, "unavailable")
	// No silent fallback to the OAuth token source.
	f.oauth.AssertNotCalled(t, "Token", context.Background())
}

func TestAuthLogout_ClearsSentinel(t *testing.T) {
	f := newAuthFixture(t)
	require.NoError(t, provider.PersistGhCliTokenSource(f.credPath))
	f.oauth.On("Logout", context.Background()).Return(nil)

	out, err := f.run(t, "auth", "logout", "--provider", "github")

	require.NoError(t, err)
	assert.Contains(t, out, "Logged out")
	f.oauth.AssertExpectations(t)
}
