package provider_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/provider"
)

func TestParseRemote_HTTPS(t *testing.T) {
	cases := []struct {
		remote string
		want   string
	}{
		{"https://github.com/owner/repo.git", "owner/repo"},
		{"https://github.com/owner/repo", "owner/repo"},
		{"https://github.com/org/my-tool.git", "org/my-tool"},
	}

	for _, tc := range cases {
		got, err := provider.ParseRemote(tc.remote)
		require.NoError(t, err, "remote: %s", tc.remote)
		assert.Equal(t, tc.want, got, "remote: %s", tc.remote)
	}
}

func TestParseRemote_SSH(t *testing.T) {
	cases := []struct {
		remote string
		want   string
	}{
		{"git@github.com:owner/repo.git", "owner/repo"},
		{"git@github.com:owner/repo", "owner/repo"},
		{"git@github.com:org/my-tool.git", "org/my-tool"},
	}

	for _, tc := range cases {
		got, err := provider.ParseRemote(tc.remote)
		require.NoError(t, err, "remote: %s", tc.remote)
		assert.Equal(t, tc.want, got, "remote: %s", tc.remote)
	}
}

func TestParseRemote_NonGitHub(t *testing.T) {
	cases := []string{
		"https://gitlab.com/owner/repo.git",
		"git@bitbucket.org:owner/repo.git",
		"https://example.com/repo.git",
	}

	for _, remote := range cases {
		_, err := provider.ParseRemote(remote)
		assert.Error(t, err, "remote: %s", remote)
	}
}
