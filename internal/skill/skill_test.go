package skill_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cdimonaco/tokenpile/internal/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList_ContainsClaudeCode(t *testing.T) {
	agents := skill.List()
	require.NotEmpty(t, agents)

	names := make([]string, len(agents))
	for i, a := range agents {
		names[i] = a.Name
	}

	assert.Contains(t, names, "claude-code")
}

func TestInstall_UnsupportedAgent(t *testing.T) {
	_, _, err := skill.Install("unknown-agent")
	assert.ErrorIs(t, err, skill.ErrUnsupportedAgent)
}

func TestInstall_WritesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, existed, err := skill.Install("claude-code")
	require.NoError(t, err)
	assert.False(t, existed)

	expected := filepath.Join(dir, ".claude", "skills", "tokenpile.md")
	data, err := os.ReadFile(expected)
	require.NoError(t, err)
	assert.Contains(t, string(data), "tokenpile log")
}

func TestInstall_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("claude-code")
	require.NoError(t, err)

	_, existed, err := skill.Install("claude-code")
	require.NoError(t, err)
	assert.True(t, existed)
}

func TestIsInstalled_False(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	assert.False(t, skill.IsInstalled("claude-code"))
}

func TestIsInstalled_True(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("claude-code")
	require.NoError(t, err)

	assert.True(t, skill.IsInstalled("claude-code"))
}
