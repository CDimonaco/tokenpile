package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/config"
	"github.com/cdimonaco/tokenpile/internal/export"
	"github.com/cdimonaco/tokenpile/internal/mocks"
	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

type resetFixture struct {
	store *store.SQLiteStore
	paths config.Paths
	auth  *mocks.AuthProvider
	priv  ed25519.PrivateKey
}

// newResetFixture builds temp config/data dirs with identity, credentials and
// pricing files, an open store with one entry, session and budget, and a mock
// auth provider. HOME is redirected so skill paths never touch the real ones.
func newResetFixture(t *testing.T) resetFixture {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgDir := filepath.Join(home, "cfg")
	dataDir := filepath.Join(home, "data")
	require.NoError(t, os.MkdirAll(cfgDir, 0o700))
	require.NoError(t, os.MkdirAll(dataDir, 0o700))

	paths := config.Paths{
		ConfigDir:       cfgDir,
		DataDir:         dataDir,
		DBPath:          filepath.Join(dataDir, "tokenpile.db"),
		PricingOverride: filepath.Join(cfgDir, "pricing.yaml"),
		IdentityKeyPath: filepath.Join(cfgDir, "identity.key"),
		IdentityPubPath: filepath.Join(cfgDir, "identity.pub"),
		CredentialsPath: filepath.Join(cfgDir, "credentials"),
	}

	priv, _, err := config.EnsureIdentity(paths)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(paths.CredentialsPath, []byte("enc"), 0o600))
	require.NoError(t, os.WriteFile(paths.PricingOverride, []byte("models: {}\n"), 0o600))

	loader, err := pricing.NewLoader("")
	require.NoError(t, err)

	s, err := store.NewSQLiteStore(paths.DBPath, loader)
	require.NoError(t, err)

	t.Cleanup(func() { _ = s.Close() })

	ctx := context.Background()
	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "o/r", IssueNum: 1, Agent: "a", Model: "m", TokensIn: 10, TokensOut: 5,
	}))

	_, err = s.StartSession(ctx, "o/r", 1)
	require.NoError(t, err)
	require.NoError(t, s.SetBudget(ctx, "o/r", 1, 12.5))

	auth := mocks.NewAuthProvider(t)
	auth.On("Token", mock.Anything).Return("", provider.ErrUnauthenticated).Maybe()
	auth.On("Logout", mock.Anything).Return(nil).Maybe()

	return resetFixture{store: s, paths: paths, auth: auth, priv: priv}
}

func runResetCmd(t *testing.T, f resetFixture, stdin string, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Reader:   strings.NewReader(stdin),
		Commands: []*cli.Command{resetCommand(f.store, f.paths, f.auth, f.priv, "test")},
	}

	err := app.RunContext(context.Background(), append([]string{"tok", "reset"}, args...))

	return buf.String(), err
}

func TestIntegration_Reset_ConfirmationAbort(t *testing.T) {
	f := newResetFixture(t)
	backup := filepath.Join(t.TempDir(), "backup.json")

	out, err := runResetCmd(t, f, "no\n", "--output", backup)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "aborted")
	assert.Contains(t, out, f.paths.DBPath)

	assert.FileExists(t, f.paths.DBPath)
	assert.FileExists(t, f.paths.IdentityKeyPath)
	assert.NoFileExists(t, backup)
}

func TestIntegration_Reset_YesResetsAndBackupVerifies(t *testing.T) {
	f := newResetFixture(t)
	backup := filepath.Join(t.TempDir(), "backup.json")

	out, err := runResetCmd(t, f, "", "--yes", "--output", backup)
	require.NoError(t, err, out)
	assert.Contains(t, out, "Reset complete.")

	for _, path := range []string{
		f.paths.DBPath, f.paths.IdentityKeyPath, f.paths.IdentityPubPath,
		f.paths.CredentialsPath, f.paths.PricingOverride,
	} {
		assert.NoFileExists(t, path)
	}

	data, err := os.ReadFile(backup)
	require.NoError(t, err)

	var doc export.Document
	require.NoError(t, json.Unmarshal(data, &doc))

	res, err := export.Verify(&doc)
	require.NoError(t, err)
	assert.False(t, res.Legacy)
	require.Len(t, doc.Entries, 1)
	require.Len(t, doc.Sessions, 1)
	require.Len(t, doc.Budgets, 1)
}

func TestIntegration_Reset_NoBackupSkipsFile(t *testing.T) {
	f := newResetFixture(t)

	out, err := runResetCmd(t, f, "", "--yes", "--no-backup")
	require.NoError(t, err, out)
	assert.NotContains(t, out, "Backup written")
	assert.NoFileExists(t, f.paths.DBPath)
}

func TestIntegration_Reset_SecondRunSucceeds(t *testing.T) {
	f := newResetFixture(t)

	_, err := runResetCmd(t, f, "", "--yes", "--no-backup")
	require.NoError(t, err)

	out, err := runResetCmd(t, f, "", "--yes", "--no-backup")
	require.NoError(t, err, out)
	assert.Contains(t, out, "Nothing to reset")
}

func TestIntegration_Reset_BackupFailureAbortsDestruction(t *testing.T) {
	f := newResetFixture(t)
	backup := filepath.Join(t.TempDir(), "missing-dir", "backup.json")

	out, err := runResetCmd(t, f, "", "--yes", "--output", backup)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "backup failed")
	assert.NotContains(t, out, "Removed:")

	assert.FileExists(t, f.paths.DBPath)
	assert.FileExists(t, f.paths.IdentityKeyPath)
	assert.FileExists(t, f.paths.CredentialsPath)
}
