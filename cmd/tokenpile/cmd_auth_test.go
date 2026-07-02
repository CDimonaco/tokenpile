package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/mocks"
)

func runAuthApp(t *testing.T, authMock *mocks.AuthProvider, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{authCommands(authMock)},
	}

	err := app.RunContext(context.Background(), append([]string{"tok"}, args...))

	return buf.String(), err
}

func TestAuthLogin_Success(t *testing.T) {
	auth := &mocks.AuthProvider{}
	auth.On("Login", context.Background()).Return(nil)

	out, err := runAuthApp(t, auth, "auth", "login", "--provider", "github")

	require.NoError(t, err)
	assert.Contains(t, out, "Authenticated")
	auth.AssertExpectations(t)
}

func TestAuthLogin_Failure(t *testing.T) {
	auth := &mocks.AuthProvider{}
	auth.On("Login", context.Background()).Return(errors.New("oauth failed"))

	_, err := runAuthApp(t, auth, "auth", "login", "--provider", "github")

	assert.Error(t, err)
	auth.AssertExpectations(t)
}

func TestAuthLogout_Success(t *testing.T) {
	auth := &mocks.AuthProvider{}
	auth.On("Logout", context.Background()).Return(nil)

	out, err := runAuthApp(t, auth, "auth", "logout", "--provider", "github")

	require.NoError(t, err)
	assert.Contains(t, out, "Logged out")
	auth.AssertExpectations(t)
}

func TestAuthStatus_Authenticated(t *testing.T) {
	auth := &mocks.AuthProvider{}
	auth.On("Token", context.Background()).Return("tok123", nil)

	out, err := runAuthApp(t, auth, "auth", "status")

	require.NoError(t, err)
	assert.True(t, strings.Contains(out, "authenticated"), "expected 'authenticated' in %q", out)
	auth.AssertExpectations(t)
}

func TestAuthStatus_NotLoggedIn(t *testing.T) {
	auth := &mocks.AuthProvider{}
	auth.On("Token", context.Background()).Return("", errors.New("not found"))

	out, err := runAuthApp(t, auth, "auth", "status")

	require.NoError(t, err)
	assert.Contains(t, out, "not logged in")
	auth.AssertExpectations(t)
}
