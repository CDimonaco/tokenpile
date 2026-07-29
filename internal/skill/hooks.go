package skill

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// hookMarker identifies tokenpile's own hook entry so it can be removed again
// without disturbing hooks the user or other tools installed.
const hookMarker = "tokenpile"

// HookPath returns where an agent's capture hook is configured, or "" when the
// agent has no hook mechanism.
func HookPath(agentName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch agentName {
	case "claude-code":
		return filepath.Join(home, ".claude", "settings.json")
	case "opencode":
		return filepath.Join(home, ".config", "opencode", "plugin", "tokenpile.js")
	default:
		return ""
	}
}

// InstallHook registers the capture hook for an agent.
//
// For Claude Code this merges a Stop hook into the user's settings file. It
// merges rather than rewrites: that file holds the user's own configuration,
// and clobbering it to add one entry would be an unacceptable trade for a usage
// tracker.
//
// Note this goes in settings, not in the skill's frontmatter. Frontmatter hooks
// are scoped to the skill's lifecycle and run only while the model has chosen
// to load the skill, which is exactly the non-determinism capture exists to
// remove.
func InstallHook(agentName string) (string, error) {
	path := HookPath(agentName)
	if path == "" {
		return "", nil
	}

	switch agentName {
	case "claude-code":
		return path, installClaudeCodeHook(path)
	case "opencode":
		return path, installOpenCodePlugin(path)
	default:
		return "", ErrUnsupportedAgent
	}
}

// UninstallHook removes what InstallHook added, leaving everything else intact.
func UninstallHook(agentName string) (string, bool, error) {
	path := HookPath(agentName)
	if path == "" {
		return "", false, nil
	}

	switch agentName {
	case "claude-code":
		removed, err := removeClaudeCodeHook(path)

		return path, removed, err
	case "opencode":
		err := os.Remove(path)
		if errors.Is(err, os.ErrNotExist) {
			return path, false, nil
		}

		if err != nil {
			return path, false, fmt.Errorf("remove plugin: %w", err)
		}

		return path, true, nil
	default:
		return "", false, ErrUnsupportedAgent
	}
}

type hookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type hookMatcher struct {
	Matcher string      `json:"matcher,omitempty"`
	Hooks   []hookEntry `json:"hooks"`
}

// readSettings loads the settings file as a generic map so every key tokenpile
// does not understand survives the round trip untouched.
func readSettings(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]any{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}

	settings := map[string]any{}
	if len(data) == 0 {
		return settings, nil
	}

	if err = json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}

	return settings, nil
}

func writeSettings(path string, settings map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create settings directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err = os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	return nil
}

func installClaudeCodeHook(path string) error {
	settings, err := readSettings(path)
	if err != nil {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}

	stop, _ := hooks["Stop"].([]any)
	stop = withoutTokenpile(stop)

	entry, err := toGeneric(hookMatcher{
		Hooks: []hookEntry{{Type: "command", Command: "tokenpile hook claude-code"}},
	})
	if err != nil {
		return err
	}

	hooks["Stop"] = append(stop, entry)
	settings["hooks"] = hooks

	return writeSettings(path, settings)
}

func removeClaudeCodeHook(path string) (bool, error) {
	settings, err := readSettings(path)
	if err != nil {
		return false, err
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		return false, nil
	}

	stop, _ := hooks["Stop"].([]any)

	pruned := withoutTokenpile(stop)
	if len(pruned) == len(stop) {
		return false, nil
	}

	if len(pruned) == 0 {
		delete(hooks, "Stop")
	} else {
		hooks["Stop"] = pruned
	}

	if len(hooks) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooks
	}

	return true, writeSettings(path, settings)
}

// withoutTokenpile drops matcher groups whose command mentions tokenpile,
// leaving every foreign hook exactly as it was.
func withoutTokenpile(groups []any) []any {
	out := make([]any, 0, len(groups))

	for _, g := range groups {
		if containsTokenpileCommand(g) {
			continue
		}

		out = append(out, g)
	}

	return out
}

func containsTokenpileCommand(group any) bool {
	m, ok := group.(map[string]any)
	if !ok {
		return false
	}

	entries, ok := m["hooks"].([]any)
	if !ok {
		return false
	}

	for _, e := range entries {
		entry, entryOK := e.(map[string]any)
		if !entryOK {
			continue
		}

		if cmd, cmdOK := entry["command"].(string); cmdOK && strings.Contains(cmd, hookMarker) {
			return true
		}
	}

	return false
}

func toGeneric(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal hook entry: %w", err)
	}

	out := map[string]any{}
	if err = json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("normalize hook entry: %w", err)
	}

	return out, nil
}

//go:embed templates/opencode-plugin.js
var openCodePlugin []byte

func installOpenCodePlugin(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create plugin directory: %w", err)
	}

	if err := os.WriteFile(path, openCodePlugin, 0o600); err != nil {
		return fmt.Errorf("write plugin: %w", err)
	}

	return nil
}
