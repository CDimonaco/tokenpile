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
	markerStart = "<!-- tokenpile:start -->"
	markerEnd   = "<!-- tokenpile:end -->"
)

var ErrUnsupportedAgent = errors.New("unsupported agent")

// Agent describes a supported coding agent and how to install the tokenpile skill into it.
type Agent struct {
	Name         string
	TemplateData []byte
	InstallPath  func() string
	// Shared indicates the install target is a file shared with other content (e.g. AGENTS.md).
	// When true, Install appends/updates a marked block instead of overwriting the whole file.
	Shared bool
}

var agents = []Agent{
	{
		Name:         "claude-code",
		TemplateData: claudeCodeTemplate,
		InstallPath: func() string {
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
		Shared:       true,
		InstallPath: func() string {
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
		Shared:       true,
		InstallPath: func() string {
			home, err := os.UserHomeDir()
			if err != nil {
				return ""
			}

			return filepath.Join(home, ".config", "opencode", "AGENTS.md")
		},
	},
}

func List() []Agent {
	out := make([]Agent, len(agents))
	copy(out, agents)

	return out
}

// Install writes the tokenpile skill for the named agent.
// For dedicated files (claude-code) it overwrites the file.
// For shared files (codex, opencode) it appends or updates a marked block.
// Returns the install path, whether the tokenpile block already existed, and any error.
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

	if agent.Shared {
		return installShared(target, agent.TemplateData)
	}

	return installDedicated(target, agent.TemplateData)
}

func installDedicated(target string, data []byte) (string, bool, error) {
	_, statErr := os.Stat(target)
	existed := statErr == nil

	if err := os.WriteFile(target, data, 0o644); err != nil { //nolint:gosec
		return "", false, fmt.Errorf("write skill file: %w", err)
	}

	return target, existed, nil
}

func installShared(target string, data []byte) (string, bool, error) {
	block := markerStart + "\n" + strings.TrimSpace(string(data)) + "\n" + markerEnd

	existing, readErr := os.ReadFile(target)
	if readErr != nil && !os.IsNotExist(readErr) {
		return "", false, fmt.Errorf("read %s: %w", target, readErr)
	}

	if os.IsNotExist(readErr) {
		if err := os.WriteFile(target, []byte(block+"\n"), 0o644); err != nil { //nolint:gosec
			return "", false, fmt.Errorf("write skill file: %w", err)
		}

		return target, false, nil
	}

	content := string(existing)
	startIdx := strings.Index(content, markerStart)
	endIdx := strings.Index(content, markerEnd)

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		updated := content[:startIdx] + block + content[endIdx+len(markerEnd):]

		if err := os.WriteFile(target, []byte(updated), 0o644); err != nil { //nolint:gosec
			return "", false, fmt.Errorf("update skill file: %w", err)
		}

		return target, true, nil
	}

	sep := "\n\n"
	if strings.HasSuffix(content, "\n\n") {
		sep = ""
	} else if strings.HasSuffix(content, "\n") {
		sep = "\n"
	}

	if err := os.WriteFile(target, []byte(content+sep+block+"\n"), 0o644); err != nil { //nolint:gosec
		return "", false, fmt.Errorf("append skill file: %w", err)
	}

	return target, false, nil
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

	if agent.Shared {
		data, err := os.ReadFile(target)
		if err != nil {
			return false
		}

		return strings.Contains(string(data), markerStart)
	}

	_, err := os.Stat(target)

	return err == nil
}

func findAgent(name string) (Agent, bool) {
	for _, a := range agents {
		if a.Name == name {
			return a, true
		}
	}

	return Agent{}, false
}
