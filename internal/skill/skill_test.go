package skill_test

import (
	"os"
	"path/filepath"
	"strings"
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

// --- claude-code (dedicated file) ---

func TestInstall_ClaudeCode_WritesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, existed, err := skill.Install("claude-code")
	require.NoError(t, err)
	assert.False(t, existed)

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "skills", "tokenpile.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "tokenpile log")
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

// --- codex (shared file, append/marker) ---

func TestInstall_Codex_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path, existed, err := skill.Install("codex")
	require.NoError(t, err)
	assert.False(t, existed)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "<!-- tokenpile:start -->")
	assert.Contains(t, content, "<!-- tokenpile:end -->")
	assert.Contains(t, content, "tokenpile log")
	assert.Contains(t, content, "--agent codex")
}

func TestInstall_Codex_AppendsToExistingFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	codexDir := filepath.Join(dir, ".codex")
	require.NoError(t, os.MkdirAll(codexDir, 0o750))

	agentsPath := filepath.Join(codexDir, "AGENTS.md")
	existing := "# My existing instructions\n\nDo stuff.\n"
	require.NoError(t, os.WriteFile(agentsPath, []byte(existing), 0o644))

	_, existed, err := skill.Install("codex")
	require.NoError(t, err)
	assert.False(t, existed)

	data, err := os.ReadFile(agentsPath)
	require.NoError(t, err)
	content := string(data)

	assert.True(t, strings.HasPrefix(content, existing), "existing content must be preserved at the top")
	assert.Contains(t, content, "<!-- tokenpile:start -->")
	assert.Contains(t, content, "--agent codex")
}

func TestInstall_Codex_UpdatesExistingBlock(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("codex")
	require.NoError(t, err)

	_, existed, err := skill.Install("codex")
	require.NoError(t, err)
	assert.True(t, existed)

	path := filepath.Join(dir, ".codex", "AGENTS.md")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.Equal(t, 1, strings.Count(string(data), "<!-- tokenpile:start -->"), "block must appear exactly once")
}

func TestIsInstalled_Codex_False(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	assert.False(t, skill.IsInstalled("codex"))
}

func TestIsInstalled_Codex_FalseWhenFileExistsWithoutMarker(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	codexDir := filepath.Join(dir, ".codex")
	require.NoError(t, os.MkdirAll(codexDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(codexDir, "AGENTS.md"), []byte("# other stuff\n"), 0o644))

	assert.False(t, skill.IsInstalled("codex"))
}

func TestIsInstalled_Codex_True(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("codex")
	require.NoError(t, err)
	assert.True(t, skill.IsInstalled("codex"))
}

// --- opencode (shared file, append/marker) ---

func TestInstall_OpenCode_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path, existed, err := skill.Install("opencode")
	require.NoError(t, err)
	assert.False(t, existed)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "<!-- tokenpile:start -->")
	assert.Contains(t, content, "--agent opencode")
}

func TestInstall_OpenCode_UpdatesExistingBlock(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("opencode")
	require.NoError(t, err)

	_, existed, err := skill.Install("opencode")
	require.NoError(t, err)
	assert.True(t, existed)

	path := filepath.Join(dir, ".config", "opencode", "AGENTS.md")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(string(data), "<!-- tokenpile:start -->"), "block must appear exactly once")
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

	skillPath := filepath.Join(dir, ".claude", "skills", "tokenpile.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0o750))

	// write a file with a stale version number
	require.NoError(t, os.WriteFile(skillPath, []byte("<!-- tokenpile-skill-version: 1 -->\n# tokenpile\n"), 0o644))

	assert.False(t, skill.IsUpToDate("claude-code"))
}

func TestIsUpToDate_NoVersionComment_False(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	skillPath := filepath.Join(dir, ".claude", "skills", "tokenpile.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(skillPath), 0o750))

	// file without any version marker (pre-v2 install)
	require.NoError(t, os.WriteFile(skillPath, []byte("# tokenpile\ntokenpile log ...\n"), 0o644))

	assert.False(t, skill.IsUpToDate("claude-code"))
}

func TestIsUpToDate_SharedAgent_AfterInstall_True(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("codex")
	require.NoError(t, err)

	assert.True(t, skill.IsUpToDate("codex"))
}

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

func TestUninstall_Codex_PreservesForeignContent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	target := filepath.Join(dir, ".codex", "AGENTS.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o750))
	require.NoError(t, os.WriteFile(target, []byte("# My agents\n\nkeep me\n"), 0o644))

	_, _, err := skill.Install("codex")
	require.NoError(t, err)

	_, removed, err := skill.Uninstall("codex")
	require.NoError(t, err)
	assert.True(t, removed)

	data, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Contains(t, string(data), "keep me")
	assert.NotContains(t, string(data), "tokenpile:start")
}

func TestUninstall_Codex_RemovesFileWhenOnlyBlock(t *testing.T) {
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
