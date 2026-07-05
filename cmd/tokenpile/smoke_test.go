package main

import (
	"os"
	"os/exec"
	"path/filepath"
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
	assert.Contains(t, string(out), "reset")
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

func TestSmoke_ExportVerify(t *testing.T) {
	bin := buildBinary(t)

	dir := t.TempDir()
	env := append(os.Environ(),
		"TOKENPILE_CONFIG_DIR="+filepath.Join(dir, "cfg"),
		"TOKENPILE_DATA_DIR="+filepath.Join(dir, "data"),
	)
	exportPath := filepath.Join(dir, "export.json")

	run := func(args ...string) (string, error) {
		cmd := exec.Command(bin, args...)
		cmd.Env = env
		out, err := cmd.CombinedOutput()

		return string(out), err
	}

	out, err := run("export", "--output", exportPath)
	require.NoError(t, err, "export failed: %s", out)

	out, err = run("export", "verify", "--file", exportPath)
	require.NoError(t, err, "verify failed: %s", out)
	assert.Contains(t, out, "OK: signature valid (schema 3.0, full document)")
	assert.Contains(t, out, "Origin not verified")

	out, err = run("export", "verify", "--file", exportPath,
		"--pubkey", filepath.Join(dir, "cfg", "identity.pub"))
	require.NoError(t, err, "verify with pubkey failed: %s", out)
	assert.Contains(t, out, "Origin verified")

	data, err := os.ReadFile(exportPath)
	require.NoError(t, err)

	tampered := strings.Replace(string(data), `"exported_by": "tokenpile/`, `"exported_by": "evil/`, 1)
	require.NotEqual(t, string(data), tampered)
	require.NoError(t, os.WriteFile(exportPath, []byte(tampered), 0o600))

	out, err = run("export", "verify", "--file", exportPath)
	require.Error(t, err, "tampered export must fail verification")
	assert.Contains(t, out, "INVALID")
}

func TestSmoke_AuthStatus(t *testing.T) {
	bin := buildBinary(t)

	// auth status should always exit 0 and print something useful
	out, err := exec.Command(bin, "auth", "status").Output()
	require.NoError(t, err)
	assert.NotEmpty(t, string(out))
}
