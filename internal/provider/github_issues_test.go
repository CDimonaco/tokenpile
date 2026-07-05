package provider_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/mocks"
	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

func newMockAuthProvider(t *testing.T, token string) *mocks.AuthProvider {
	t.Helper()
	m := mocks.NewAuthProvider(t)
	m.On("Token", context.Background()).Return(token, nil)

	return m
}

func TestGitHubIssueProvider_ListIssues(t *testing.T) {
	issues := []map[string]any{
		{"number": 1, "title": "Bug fix", "state": "open", "html_url": "https://github.com/o/r/issues/1"},
		{"number": 2, "title": "Feature", "state": "open", "html_url": "https://github.com/o/r/issues/2"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/repos/o/r/issues", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(issues)
	}))
	defer srv.Close()

	auth := newMockAuthProvider(t, "test-token")
	p := provider.NewGitHubIssueProviderWithURL(auth, srv.URL)

	got, err := p.ListIssues(context.Background(), usage.Filter{Repo: "o/r", State: "open"})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, 1, got[0].Number)
	assert.Equal(t, "Bug fix", got[0].Title)
}

func TestGitHubIssueProvider_ListIssues_Paginated(t *testing.T) {
	pageOne := []map[string]any{
		{"number": 1, "title": "First", "state": "open", "html_url": "https://github.com/o/r/issues/1"},
	}
	pageTwo := []map[string]any{
		{"number": 2, "title": "Second", "state": "open", "html_url": "https://github.com/o/r/issues/2"},
	}

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/repos/o/r/issues", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode(pageTwo)
			return
		}

		w.Header().Set("Link",
			`<`+srv.URL+`/api/v3/repos/o/r/issues?page=2>; rel="next", <`+srv.URL+`/api/v3/repos/o/r/issues?page=2>; rel="last"`)
		_ = json.NewEncoder(w).Encode(pageOne)
	}))
	defer srv.Close()

	auth := newMockAuthProvider(t, "test-token")
	p := provider.NewGitHubIssueProviderWithURL(auth, srv.URL)

	got, err := p.ListIssues(context.Background(), usage.Filter{Repo: "o/r", State: "open"})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, 1, got[0].Number)
	assert.Equal(t, 2, got[1].Number)
}

func TestGitHubIssueProvider_ListIssues_UnauthenticatedError(t *testing.T) {
	m := mocks.NewAuthProvider(t)
	m.On("Token", context.Background()).Return("", provider.ErrUnauthenticated)

	p := provider.NewGitHubIssueProvider(m)

	_, err := p.ListIssues(context.Background(), usage.Filter{Repo: "o/r"})
	assert.ErrorIs(t, err, provider.ErrUnauthenticated)
}

func TestGitHubIssueProvider_GetIssue(t *testing.T) {
	issue := map[string]any{
		"number":   42,
		"title":    "Critical bug",
		"state":    "open",
		"html_url": "https://github.com/o/r/issues/42",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/repos/o/r/issues/42", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(issue)
	}))
	defer srv.Close()

	auth := newMockAuthProvider(t, "test-token")
	p := provider.NewGitHubIssueProviderWithURL(auth, srv.URL)

	got, err := p.GetIssue(context.Background(), "o/r", 42)
	require.NoError(t, err)
	assert.Equal(t, 42, got.Number)
	assert.Equal(t, "Critical bug", got.Title)
}
