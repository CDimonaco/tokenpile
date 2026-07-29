package provider_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/provider"
)

func userServer(t *testing.T, scopes string, hasScopeHeader bool) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if hasScopeHeader {
			w.Header().Set("X-OAuth-Scopes", scopes)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"login":"octocat"}`))
	}))
	t.Cleanup(srv.Close)

	return srv
}

func TestValidateToken_ReadsLoginAndScopes(t *testing.T) {
	srv := userServer(t, "read:user, repo", true)

	info, err := provider.ValidateTokenWithURL(context.Background(), "tok", srv.URL)

	require.NoError(t, err)
	assert.Equal(t, "octocat", info.Login)
	assert.True(t, info.ScopesKnown)
	assert.Equal(t, []string{"read:user", "repo"}, info.Scopes)
	assert.True(t, info.HasScope("repo"))
}

func TestValidateToken_MissingScopeDetected(t *testing.T) {
	srv := userServer(t, "read:user", true)

	info, err := provider.ValidateTokenWithURL(context.Background(), "tok", srv.URL)

	require.NoError(t, err)
	assert.True(t, info.ScopesKnown)
	assert.False(t, info.HasScope("repo"))
}

// Fine-grained PATs send no X-OAuth-Scopes header. Absence must not be read as
// an empty scope set, or every fine-grained PAT would be rejected.
func TestValidateToken_AbsentHeaderMeansScopesUnknown(t *testing.T) {
	srv := userServer(t, "", false)

	info, err := provider.ValidateTokenWithURL(context.Background(), "tok", srv.URL)

	require.NoError(t, err)
	assert.False(t, info.ScopesKnown)
	assert.True(t, info.HasScope("repo"), "an unknown scope set cannot be shown to be insufficient")
}

func TestValidateToken_EmptyHeaderIsAKnownEmptyScopeSet(t *testing.T) {
	srv := userServer(t, "", true)

	info, err := provider.ValidateTokenWithURL(context.Background(), "tok", srv.URL)

	require.NoError(t, err)
	assert.True(t, info.ScopesKnown)
	assert.False(t, info.HasScope("repo"))
}

func TestValidateToken_RejectsBadCredential(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
	}))
	t.Cleanup(srv.Close)

	_, err := provider.ValidateTokenWithURL(context.Background(), "tok", srv.URL)

	require.Error(t, err)
}
