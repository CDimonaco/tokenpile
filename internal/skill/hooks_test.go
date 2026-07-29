package skill_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/skill"
)

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	out := map[string]any{}
	require.NoError(t, json.Unmarshal(data, &out))

	return out
}

func TestInstallHook_ClaudeCode_CreatesStopHook(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path, err := skill.InstallHook("claude-code")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, ".claude", "settings.json"), path)

	settings := readJSON(t, path)
	hooks, ok := settings["hooks"].(map[string]any)
	require.True(t, ok)
	stop, ok := hooks["Stop"].([]any)
	require.True(t, ok)
	require.Len(t, stop, 1)
}

// The settings file is the user's own configuration. Adding one hook must not
// cost them anything else in it.
func TestInstallHook_PreservesForeignSettingsAndHooks(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path := filepath.Join(dir, ".claude", "settings.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))

	existing := `{
	  "model": "opus",
	  "permissions": {"allow": ["Bash(ls)"]},
	  "hooks": {
	    "Stop": [{"hooks": [{"type": "command", "command": "notify-send done"}]}],
	    "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "audit.sh"}]}]
	  }
	}`
	require.NoError(t, os.WriteFile(path, []byte(existing), 0o600))

	_, err := skill.InstallHook("claude-code")
	require.NoError(t, err)

	settings := readJSON(t, path)
	assert.Equal(t, "opus", settings["model"])
	assert.NotNil(t, settings["permissions"])

	hooks := settings["hooks"].(map[string]any)
	assert.NotNil(t, hooks["PreToolUse"], "foreign event untouched")

	stop := hooks["Stop"].([]any)
	require.Len(t, stop, 2, "the user's own Stop hook survives alongside tokenpile's")
}

func TestUninstallHook_RemovesOnlyTokenpile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path := filepath.Join(dir, ".claude", "settings.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte(
		`{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"notify-send done"}]}]}}`), 0o600))

	_, err := skill.InstallHook("claude-code")
	require.NoError(t, err)

	_, removed, err := skill.UninstallHook("claude-code")
	require.NoError(t, err)
	assert.True(t, removed)

	settings := readJSON(t, path)
	hooks := settings["hooks"].(map[string]any)
	stop := hooks["Stop"].([]any)

	require.Len(t, stop, 1, "the user's hook remains")

	entry := stop[0].(map[string]any)["hooks"].([]any)[0].(map[string]any)
	assert.Equal(t, "notify-send done", entry["command"])
}

func TestInstallHook_IsIdempotent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, err := skill.InstallHook("claude-code")
	require.NoError(t, err)
	_, err = skill.InstallHook("claude-code")
	require.NoError(t, err)

	settings := readJSON(t, filepath.Join(dir, ".claude", "settings.json"))
	stop := settings["hooks"].(map[string]any)["Stop"].([]any)

	assert.Len(t, stop, 1, "reinstalling must not stack duplicate hooks")
}

func TestUninstallHook_NotInstalled(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, removed, err := skill.UninstallHook("claude-code")

	require.NoError(t, err)
	assert.False(t, removed)
}

func TestInstallHook_OpenCode_WritesPlugin(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path, err := skill.InstallHook("opencode")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, ".config", "opencode", "plugin", "tokenpile.js"), path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "session.idle")
	assert.Contains(t, string(data), "tokenpile hook opencode")
}

func TestInstall_AlsoInstallsHook(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_, _, err := skill.Install("claude-code")
	require.NoError(t, err)

	settings := readJSON(t, filepath.Join(dir, ".claude", "settings.json"))
	assert.NotNil(t, settings["hooks"], "capture is inert without the hook")
}

// Token counts come from the transcript now. A template that still asks the
// model for them would reintroduce the estimate this whole change removes.
func TestTemplates_DoNotAskForTokenCounts(t *testing.T) {
	for _, agent := range skill.List() {
		t.Run(agent.Name, func(t *testing.T) {
			content := string(agent.TemplateData)

			assert.NotContains(t, content, "tokenpile log")
			assert.NotContains(t, content, "--tokens-in")
			assert.NotContains(t, content, "--input")
			assert.Contains(t, content, "tokenpile bind")
		})
	}
}
