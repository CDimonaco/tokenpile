package provider_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/provider"
)

// ghStubDir writes a stub gh executable into a temp dir and returns the dir, so
// callers can put it on PATH. The script appends a line to a call log on every
// invocation, letting tests assert that the provider shells out each time
// instead of caching.
func ghStubDir(t *testing.T, script string) string {
	t.Helper()

	dir := t.TempDir()

	body := fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %q\n%s\n", ghCallLog(dir), script)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gh"), []byte(body), 0o700))

	return dir
}

func ghCallLog(dir string) string {
	return filepath.Join(dir, "calls.log")
}

// fakeGh installs the stub as the only entry on PATH and returns its call log.
func fakeGh(t *testing.T, script string) string {
	t.Helper()

	dir := ghStubDir(t, script)
	t.Setenv("PATH", dir)

	return ghCallLog(dir)
}

func callCount(t *testing.T, callLog string) int {
	t.Helper()

	data, err := os.ReadFile(callLog)
	if os.IsNotExist(err) {
		return 0
	}

	require.NoError(t, err)

	count := 0

	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	return count
}

func TestGhCliToken_ReturnsTrimmedStdout(t *testing.T) {
	fakeGh(t, `echo "  gho_testtoken  "`)

	tok, err := provider.NewGhCliAuthProvider().Token(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "gho_testtoken", tok)
}

func TestGhCliToken_IsNotCached(t *testing.T) {
	callLog := fakeGh(t, `echo gho_testtoken`)

	p := provider.NewGhCliAuthProvider()

	_, err := p.Token(context.Background())
	require.NoError(t, err)

	_, err = p.Token(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 2, callCount(t, callLog), "gh must be executed on every Token call")
}

func TestGhCliToken_MissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := provider.NewGhCliAuthProvider().Token(context.Background())

	require.ErrorIs(t, err, provider.ErrGhNotFound)
	require.NotErrorIs(t, err, provider.ErrUnauthenticated)
}

func TestGhCliToken_NotAuthenticated(t *testing.T) {
	fakeGh(t, `echo "gh: not logged in" >&2; exit 1`)

	_, err := provider.NewGhCliAuthProvider().Token(context.Background())

	require.ErrorIs(t, err, provider.ErrGhUnauthenticated)
	require.NotErrorIs(t, err, provider.ErrUnauthenticated)
	assert.Contains(t, err.Error(), "not logged in", "gh's own diagnostic should survive")
}

func TestGhCliToken_EmptyOutputIsUnauthenticated(t *testing.T) {
	fakeGh(t, `echo ""`)

	_, err := provider.NewGhCliAuthProvider().Token(context.Background())

	require.ErrorIs(t, err, provider.ErrGhUnauthenticated)
}

func TestGhCliGhLogin(t *testing.T) {
	fakeGh(t, `echo octocat`)

	login, err := provider.NewGhCliAuthProvider().GhLogin(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "octocat", login)
}

func TestGhCliAvailable(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   bool
	}{
		{name: "authenticated", script: `echo gho_testtoken`, want: true},
		{name: "not authenticated", script: `exit 1`, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeGh(t, tc.script)

			got := provider.NewGhCliAuthProvider().Available(context.Background())

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGhCliAvailable_MissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	assert.False(t, provider.NewGhCliAuthProvider().Available(context.Background()))
}

// Logout must never touch gh's own credential: tokenpile borrows it, it does
// not own it.
func TestGhCliLogout_LeavesGhCredentialAlone(t *testing.T) {
	callLog := fakeGh(t, `echo gho_testtoken`)

	require.NoError(t, provider.NewGhCliAuthProvider().Logout(context.Background()))

	assert.Equal(t, 0, callCount(t, callLog))
}

func TestGhCliLogin_FailsWhenGhUnavailable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := provider.NewGhCliAuthProvider().Login(context.Background())

	require.ErrorIs(t, err, provider.ErrGhNotFound)
}
