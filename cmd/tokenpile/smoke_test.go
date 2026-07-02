package main

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildBinary(t *testing.T) string {
	t.Helper()

	bin := t.TempDir() + "/tokenpile"
	cmd := exec.Command("go", "build", "-o", bin, ".")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "build failed: %s", out)

	return bin
}

func TestSmoke_Version(t *testing.T) {
	bin := buildBinary(t)

	out, err := exec.Command(bin, "--version").Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "tokenpile")
}

func TestSmoke_HelpFlag(t *testing.T) {
	bin := buildBinary(t)

	out, err := exec.Command(bin, "--help").Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "log")
	assert.Contains(t, string(out), "report")
	assert.Contains(t, string(out), "auth")
	assert.Contains(t, string(out), "export")
	assert.Contains(t, string(out), "skill")
}

func TestSmoke_PricingList(t *testing.T) {
	bin := buildBinary(t)

	out, err := exec.Command(bin, "pricing", "list").Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "claude")
}

func TestSmoke_SkillList(t *testing.T) {
	bin := buildBinary(t)

	out, err := exec.Command(bin, "skill", "list").Output()
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(string(out)), "claude-code")
}

func TestSmoke_LogMissingFlags(t *testing.T) {
	bin := buildBinary(t)

	cmd := exec.Command(bin, "log", "--repo", "owner/repo")
	err := cmd.Run()
	assert.Error(t, err, "log without required flags should fail")
}

func TestSmoke_AuthStatus(t *testing.T) {
	bin := buildBinary(t)

	// auth status should always exit 0 and print something useful
	out, err := exec.Command(bin, "auth", "status").Output()
	require.NoError(t, err)
	assert.NotEmpty(t, string(out))
}
