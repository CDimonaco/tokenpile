package provider_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"

	"github.com/cdimonaco/tokenpile/internal/provider"
)

// newCredPath isolates each test from the real OS keychain and from every other
// test's credential slot.
func newCredPath(t *testing.T) string {
	t.Helper()
	keyring.MockInit()

	return filepath.Join(t.TempDir(), "credentials")
}

func TestStoredTokenSource_EmptySlotIsOAuth(t *testing.T) {
	credPath := newCredPath(t)

	assert.Equal(t, provider.TokenSourceOAuth, provider.StoredTokenSource(credPath))
}

func TestStoredTokenSource_SentinelRoundTrips(t *testing.T) {
	credPath := newCredPath(t)

	require.NoError(t, provider.PersistGhCliTokenSource(credPath))

	assert.Equal(t, provider.TokenSourceGhCli, provider.StoredTokenSource(credPath))
}

func TestStoredTokenSource_RealTokenIsOAuth(t *testing.T) {
	credPath := newCredPath(t)

	require.NoError(t, keyring.Set("tokenpile", "github-token", "gho_realtoken"))

	assert.Equal(t, provider.TokenSourceOAuth, provider.StoredTokenSource(credPath))
}

// The sentinel must never reach GitHub as a bearer token.
func TestGitHubAuthProvider_SentinelIsNotAToken(t *testing.T) {
	credPath := newCredPath(t)

	require.NoError(t, provider.PersistGhCliTokenSource(credPath))

	auth := provider.NewGitHubAuthProvider("id", "secret", credPath)

	tok, err := auth.Token(context.Background())

	require.ErrorIs(t, err, provider.ErrUnauthenticated)
	assert.Empty(t, tok)
}

func TestGitHubAuthProvider_RealTokenIsReturned(t *testing.T) {
	credPath := newCredPath(t)

	require.NoError(t, keyring.Set("tokenpile", "github-token", "gho_realtoken"))

	auth := provider.NewGitHubAuthProvider("id", "secret", credPath)

	tok, err := auth.Token(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "gho_realtoken", tok)
}

// Logout owns the slot regardless of which source is stored in it: leaving the
// sentinel behind would keep a reset or logged-out machine pointed at gh.
func TestLogout_ClearsSentinel(t *testing.T) {
	credPath := newCredPath(t)

	require.NoError(t, provider.PersistGhCliTokenSource(credPath))
	require.Equal(t, provider.TokenSourceGhCli, provider.StoredTokenSource(credPath))

	auth := provider.NewGitHubAuthProvider("id", "secret", credPath)
	require.NoError(t, auth.Logout(context.Background()))

	assert.Equal(t, provider.TokenSourceOAuth, provider.StoredTokenSource(credPath))

	tok, err := auth.Token(context.Background())
	require.ErrorIs(t, err, provider.ErrUnauthenticated)
	assert.Empty(t, tok)
}

func TestIsGhCredentialError(t *testing.T) {
	assert.True(t, provider.IsGhCredentialError(provider.ErrGhNotFound))
	assert.True(t, provider.IsGhCredentialError(provider.ErrGhUnauthenticated))
	assert.False(t, provider.IsGhCredentialError(provider.ErrUnauthenticated))
}
