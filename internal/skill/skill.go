package skill

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/claude-code.md
var claudeCodeTemplate []byte

//go:embed templates/codex.md
var codexTemplate []byte

//go:embed templates/opencode.md
var opencodeTemplate []byte

const (
	markerStart         = "<!-- tokenpile:start -->"
	markerEnd           = "<!-- tokenpile:end -->"
	versionMarkerPrefix = "<!-- tokenpile-skill-version:"
)

var ErrUnsupportedAgent = errors.New("unsupported agent")

// Agent describes a supported coding agent and how to install the tokenpile skill into it.
type Agent struct {
	Name         string
	TemplateData []byte
	InstallPath  func() string
	// LegacyDedicatedPath, when set, is a previous dedicated-file install
	// location that Install/Uninstall removes outright on a best-effort basis.
	LegacyDedicatedPath func() string
	// LegacySharedPath, when set, is a previous shared-file install location
	// (e.g. an AGENTS.md with a marked tokenpile block) that Install/Uninstall
	// strips on a best-effort basis, leaving the rest of the file untouched.
	LegacySharedPath func() string
}

var agents = []Agent{
	{
		Name:         "claude-code",
		TemplateData: claudeCodeTemplate,
		InstallPath: func() string {
			return skillPath(".claude")
		},
		LegacyDedicatedPath: func() string {
			home, err := os.UserHomeDir()
			if err != nil {
				return ""
			}

			return filepath.Join(home, ".claude", "skills", "tokenpile.md")
		},
	},
	{
		Name:         "codex",
		TemplateData: codexTemplate,
		InstallPath: func() string {
			return skillPath(".codex")
		},
		LegacySharedPath: func() string {
			home, err := os.UserHomeDir()
			if err != nil {
				return ""
			}

			return filepath.Join(home, ".codex", "AGENTS.md")
		},
	},
	{
		Name:         "opencode",
		TemplateData: opencodeTemplate,
		InstallPath: func() string {
			home, err := os.UserHomeDir()
			if err != nil {
				return ""
			}

			return filepath.Join(home, ".config", "opencode", "skills", "tokenpile", "SKILL.md")
		},
		LegacySharedPath: func() string {
			home, err := os.UserHomeDir()
			if err != nil {
				return ""
			}

			return filepath.Join(home, ".config", "opencode", "AGENTS.md")
		},
	},
}

// skillPath builds the Agent Skills Spec layout (~/<dir>/skills/tokenpile/SKILL.md)
// used by both Claude Code and OpenCode's compatible-path discovery.
func skillPath(dotDir string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, dotDir, "skills", "tokenpile", "SKILL.md")
}

func List() []Agent {
	out := make([]Agent, len(agents))
	copy(out, agents)

	return out
}

// Install writes the tokenpile skill for the named agent as a dedicated
// SKILL.md file, cleaning up any previous install location on a best-effort
// basis. Returns the install path, whether the file already existed, and any error.
func Install(agentName string) (string, bool, error) {
	agent, found := findAgent(agentName)
	if !found {
		return "", false, fmt.Errorf("%w: %s", ErrUnsupportedAgent, agentName)
	}

	target := agent.InstallPath()
	if target == "" {
		return "", false, fmt.Errorf("cannot determine install path for agent %s", agentName)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		return "", false, fmt.Errorf("create skill directory: %w", err)
	}

	cleanupLegacy(agent, target)

	return installDedicated(target, agent.TemplateData)
}

// cleanupLegacy removes stale installs from formats predating the SKILL.md
// migration, best-effort: failures here must never block a fresh install.
func cleanupLegacy(agent Agent, target string) {
	if agent.LegacyDedicatedPath != nil {
		if legacy := agent.LegacyDedicatedPath(); legacy != "" && legacy != target {
			_ = os.Remove(legacy)
		}
	}

	if agent.LegacySharedPath != nil {
		if legacy := agent.LegacySharedPath(); legacy != "" {
			_, _, _ = uninstallShared(legacy)
		}
	}
}

func installDedicated(target string, data []byte) (string, bool, error) {
	_, statErr := os.Stat(target)
	existed := statErr == nil

	if err := os.WriteFile(target, data, 0o644); err != nil { //nolint:gosec
		return "", false, fmt.Errorf("write skill file: %w", err)
	}

	return target, existed, nil
}

// Uninstall reverses Install for the named agent: the dedicated SKILL.md file
// is removed, and any leftover legacy install (pre-migration flat file or
// AGENTS.md block) is cleaned up on a best-effort basis. Uninstalling an
// agent with no installed skill succeeds and reports that nothing was removed.
// Returns the install path, whether anything was removed, and any error.
func Uninstall(agentName string) (string, bool, error) {
	agent, found := findAgent(agentName)
	if !found {
		return "", false, fmt.Errorf("%w: %s", ErrUnsupportedAgent, agentName)
	}

	target := agent.InstallPath()
	if target == "" {
		return "", false, fmt.Errorf("cannot determine install path for agent %s", agentName)
	}

	cleanupLegacy(agent, target)

	return uninstallDedicated(target)
}

func uninstallDedicated(target string) (string, bool, error) {
	if err := os.Remove(target); err != nil {
		if os.IsNotExist(err) {
			return target, false, nil
		}

		return target, false, fmt.Errorf("remove skill file: %w", err)
	}

	return target, true, nil
}

// uninstallShared strips the marked tokenpile block from a shared file (used
// only to clean up pre-migration AGENTS.md installs), removing the file
// entirely when nothing else remains.
func uninstallShared(target string) (string, bool, error) {
	existing, err := os.ReadFile(target)
	if err != nil {
		if os.IsNotExist(err) {
			return target, false, nil
		}

		return target, false, fmt.Errorf("read %s: %w", target, err)
	}

	content := string(existing)
	startIdx := strings.Index(content, markerStart)
	endIdx := strings.Index(content, markerEnd)

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return target, false, nil
	}

	remaining := content[:startIdx] + content[endIdx+len(markerEnd):]

	if strings.TrimSpace(remaining) == "" {
		if err = os.Remove(target); err != nil {
			return target, false, fmt.Errorf("remove skill file: %w", err)
		}

		return target, true, nil
	}

	remaining = strings.TrimRight(remaining, "\n") + "\n"

	if err = os.WriteFile(target, []byte(remaining), 0o644); err != nil { //nolint:gosec
		return target, false, fmt.Errorf("update skill file: %w", err)
	}

	return target, true, nil
}

func IsInstalled(agentName string) bool {
	agent, found := findAgent(agentName)
	if !found {
		return false
	}

	target := agent.InstallPath()
	if target == "" {
		return false
	}

	_, err := os.Stat(target)

	return err == nil
}

// IsUpToDate reports whether the installed skill file matches the embedded
// template version. Returns false when the agent is not installed or the
// installed file predates the version comment added in v2.
func IsUpToDate(agentName string) bool {
	agent, found := findAgent(agentName)
	if !found {
		return false
	}

	target := agent.InstallPath()
	if target == "" {
		return false
	}

	installedVersion := extractVersionFromFile(target)
	embeddedVersion := extractVersionFromBytes(agent.TemplateData)

	return installedVersion != "" && installedVersion == embeddedVersion
}

func extractVersionFromFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return extractVersionFromBytes(data)
}

func extractVersionFromBytes(data []byte) string {
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, versionMarkerPrefix); ok {
			ver := strings.TrimSuffix(after, "-->")

			return strings.TrimSpace(ver)
		}
	}

	return ""
}

func findAgent(name string) (Agent, bool) {
	for _, a := range agents {
		if a.Name == name {
			return a, true
		}
	}

	return Agent{}, false
}
