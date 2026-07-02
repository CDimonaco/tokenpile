package skill

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed templates/claude-code.md
var claudeCodeTemplate []byte

var ErrUnsupportedAgent = errors.New("unsupported agent")

type Agent struct {
	Name         string
	TemplateData []byte
	InstallPath  func() string
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
}

func List() []Agent {
	out := make([]Agent, len(agents))
	copy(out, agents)

	return out
}

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

	_, statErr := os.Stat(target)
	existed := statErr == nil

	if err := os.WriteFile(target, agent.TemplateData, 0o644); err != nil { //nolint:gosec
		return "", false, fmt.Errorf("write skill file: %w", err)
	}

	return target, existed, nil
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

func findAgent(name string) (Agent, bool) {
	for _, a := range agents {
		if a.Name == name {
			return a, true
		}
	}

	return Agent{}, false
}
