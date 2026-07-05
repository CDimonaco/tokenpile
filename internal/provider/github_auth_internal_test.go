package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptedTokenRoundtrip(t *testing.T) {
	credPath := filepath.Join(t.TempDir(), "credentials")
	p := NewGitHubAuthProvider("id", "secret", credPath)

	require.NoError(t, p.storeEncryptedToken("gho_testtoken"))

	got, err := p.loadEncryptedToken()
	require.NoError(t, err)
	assert.Equal(t, "gho_testtoken", got)
}

func TestLoadEncryptedToken_CorruptedFile(t *testing.T) {
	credPath := filepath.Join(t.TempDir(), "credentials")
	p := NewGitHubAuthProvider("id", "secret", credPath)

	require.NoError(t, p.storeEncryptedToken("gho_testtoken"))

	data, err := os.ReadFile(credPath)
	require.NoError(t, err)

	data[len(data)-1] ^= 0xff
	require.NoError(t, os.WriteFile(credPath, data, 0o600))

	_, err = p.loadEncryptedToken()
	require.Error(t, err)
}

func TestMachineKey_Deterministic(t *testing.T) {
	first := machineKey()
	second := machineKey()

	assert.Equal(t, first, second)
	assert.Len(t, first, 32)
}
