package skill_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/skill"
)

func TestList_ContainsAllAgents(t *testing.T) {
	agents := skill.List()
	names := make([]string, len(agents))

	for i, a := range agents {
		names[i] = a.Name
	}

	assert.Contains(t, names, "claude-code")
	assert.Contains(t, names, "codex")
	assert.Contains(t, names, "opencode")
}

func TestInstall_UnsupportedAgent(t *testing.T) {
	_, _, err := skill.Install("unknown-agent")
	assert.ErrorIs(t, err, skill.ErrUnsupportedAgent)
}

// --- claude-code ---

func TestInstall_ClaudeCode_WritesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, existed, err := skill.Install("claude-code")
	require.NoError(t, err)
	assert.False(t, existed)

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "skills", "tokenpile", "SKILL.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "tokenpile log")
	assert.Contains(t, string(data), "name: tokenpile")
}

func TestInstall_ClaudeCode_RemovesLegacyFlatFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	legacyPath := filepath.Join(dir, ".claude", "skills", "tokenpile.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(legacyPath), 0o750))
	require.NoError(t, os.WriteFile(legacyPath, []byte("old flat skill\n"), 0o644))

	_, _, err := skill.Install("claude-code")
	require.NoError(t, err)

	_, statErr := os.Stat(legacyPath)
	assert.True(t, os.IsNotExist(statErr), "legacy flat file should be removed on install")
}

func TestInstall_ClaudeCode_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("claude-code")
	require.NoError(t, err)

	_, existed, err := skill.Install("claude-code")
	require.NoError(t, err)
	assert.True(t, existed)
}

func TestIsInstalled_ClaudeCode_False(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	assert.False(t, skill.IsInstalled("claude-code"))
}

func TestIsInstalled_ClaudeCode_True(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("claude-code")
	require.NoError(t, err)
	assert.True(t, skill.IsInstalled("claude-code"))
}

// --- codex ---

func TestInstall_Codex_WritesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path, existed, err := skill.Install("codex")
	require.NoError(t, err)
	assert.False(t, existed)
	assert.Equal(t, filepath.Join(dir, ".codex", "skills", "tokenpile", "SKILL.md"), path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "name: tokenpile")
	assert.Contains(t, content, "tokenpile log")
	assert.Contains(t, content, "--agent codex")
}

func TestInstall_Codex_RemovesLegacyAgentsBlockButKeepsForeignContent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	agentsPath := filepath.Join(dir, ".codex", "AGENTS.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentsPath), 0o750))
	existing := "# My existing instructions\n\n<!-- tokenpile:start -->\nold block\n<!-- tokenpile:end -->\n"
	require.NoError(t, os.WriteFile(agentsPath, []byte(existing), 0o644))

	_, _, err := skill.Install("codex")
	require.NoError(t, err)

	data, err := os.ReadFile(agentsPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "My existing instructions")
	assert.NotContains(t, content, "tokenpile:start")
}

func TestInstall_Codex_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("codex")
	require.NoError(t, err)

	_, existed, err := skill.Install("codex")
	require.NoError(t, err)
	assert.True(t, existed)
}

func TestIsInstalled_Codex_False(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	assert.False(t, skill.IsInstalled("codex"))
}

func TestIsInstalled_Codex_True(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("codex")
	require.NoError(t, err)
	assert.True(t, skill.IsInstalled("codex"))
}

// --- opencode ---

func TestInstall_OpenCode_WritesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path, existed, err := skill.Install("opencode")
	require.NoError(t, err)
	assert.False(t, existed)
	assert.Equal(t, filepath.Join(dir, ".config", "opencode", "skills", "tokenpile", "SKILL.md"), path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "name: tokenpile")
	assert.Contains(t, content, "--agent opencode")
}

func TestInstall_OpenCode_RemovesLegacyAgentsBlockButKeepsForeignContent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	agentsPath := filepath.Join(dir, ".config", "opencode", "AGENTS.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentsPath), 0o750))
	existing := "# keep me\n\n<!-- tokenpile:start -->\nold block\n<!-- tokenpile:end -->\n"
	require.NoError(t, os.WriteFile(agentsPath, []byte(existing), 0o644))

	_, _, err := skill.Install("opencode")
	require.NoError(t, err)

	data, err := os.ReadFile(agentsPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "keep me")
	assert.NotContains(t, content, "tokenpile:start")
}

func TestInstall_OpenCode_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("opencode")
	require.NoError(t, err)

	_, existed, err := skill.Install("opencode")
	require.NoError(t, err)
	assert.True(t, existed)
}

// --- IsUpToDate ---

func TestIsUpToDate_UnknownAgent_False(t *testing.T) {
	assert.False(t, skill.IsUpToDate("no-such-agent"))
}

func TestIsUpToDate_NotInstalled_False(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	assert.False(t, skill.IsUpToDate("claude-code"))
}

func TestIsUpToDate_AfterInstall_True(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("claude-code")
	require.NoError(t, err)

	assert.True(t, skill.IsUpToDate("claude-code"))
}

func TestIsUpToDate_OutdatedFile_False(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path := filepath.Join(dir, ".claude", "skills", "tokenpile", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))

	// write a file with a stale version number
	require.NoError(t, os.WriteFile(path, []byte("<!-- tokenpile-skill-version: 1 -->\n# tokenpile\n"), 0o644))

	assert.False(t, skill.IsUpToDate("claude-code"))
}

func TestIsUpToDate_NoVersionComment_False(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path := filepath.Join(dir, ".claude", "skills", "tokenpile", "SKILL.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))

	// file without any version marker (pre-v2 install)
	require.NoError(t, os.WriteFile(path, []byte("# tokenpile\ntokenpile log ...\n"), 0o644))

	assert.False(t, skill.IsUpToDate("claude-code"))
}

func TestIsUpToDate_Codex_AfterInstall_True(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("codex")
	require.NoError(t, err)

	assert.True(t, skill.IsUpToDate("codex"))
}

// --- Uninstall ---

func TestUninstall_ClaudeCode_RemovesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path, _, err := skill.Install("claude-code")
	require.NoError(t, err)

	gotPath, removed, err := skill.Uninstall("claude-code")
	require.NoError(t, err)
	assert.True(t, removed)
	assert.Equal(t, path, gotPath)

	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr))
}

func TestUninstall_Codex_RemovesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path, _, err := skill.Install("codex")
	require.NoError(t, err)

	_, removed, err := skill.Uninstall("codex")
	require.NoError(t, err)
	assert.True(t, removed)

	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr))
}

func TestUninstall_Codex_CleansLegacyAgentsBlockToo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	agentsPath := filepath.Join(dir, ".codex", "AGENTS.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentsPath), 0o750))
	existing := "# keep me\n\n<!-- tokenpile:start -->\nold block\n<!-- tokenpile:end -->\n"
	require.NoError(t, os.WriteFile(agentsPath, []byte(existing), 0o644))

	_, _, err := skill.Install("codex")
	require.NoError(t, err)

	_, _, err = skill.Uninstall("codex")
	require.NoError(t, err)

	data, err := os.ReadFile(agentsPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "keep me")
	assert.NotContains(t, content, "tokenpile:start")
}

func TestUninstall_NotInstalled_NoOp(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, removed, err := skill.Uninstall("claude-code")
	require.NoError(t, err)
	assert.False(t, removed)

	_, removed, err = skill.Uninstall("codex")
	require.NoError(t, err)
	assert.False(t, removed)
}

func TestUninstall_UnsupportedAgent(t *testing.T) {
	_, _, err := skill.Uninstall("unknown")
	assert.ErrorIs(t, err, skill.ErrUnsupportedAgent)
}
