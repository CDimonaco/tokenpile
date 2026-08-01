package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/attribution"
	"github.com/cdimonaco/tokenpile/internal/capture"
	"github.com/cdimonaco/tokenpile/internal/config"
	"github.com/cdimonaco/tokenpile/internal/pricing"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

func hookPaths(t *testing.T) config.Paths {
	t.Helper()

	dir := t.TempDir()

	return config.Paths{
		BindingsPath: filepath.Join(dir, "bindings.json"),
		SpoolPath:    filepath.Join(dir, "spool.jsonl"),
	}
}

func runHook(t *testing.T, paths config.Paths, stdin string, args ...string) error {
	t.Helper()

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Reader:   strings.NewReader(stdin),
		Commands: []*cli.Command{hookCommand(paths)},
	}

	return app.RunContext(context.Background(), append([]string{"tok", "hook"}, args...))
}

const hookTranscript = `{"type":"assistant","uuid":"a1","sessionId":"s1","cwd":"/w","gitBranch":"feat/42-x","timestamp":"2026-07-29T10:00:00Z","message":{"role":"assistant","model":"claude-opus-5","usage":{"input_tokens":5,"cache_creation_input_tokens":100,"cache_read_input_tokens":9000,"output_tokens":50}}}
`

// The hook writes to the spool, never straight to the database: a capture that
// cannot reach storage must still leave a durable record.
func TestHook_ClaudeCode_SpoolsTurns(t *testing.T) {
	paths := hookPaths(t)

	transcript := filepath.Join(t.TempDir(), "transcript.jsonl")
	require.NoError(t, os.WriteFile(transcript, []byte(hookTranscript), 0o600))

	payload := `{"session_id":"s1","transcript_path":"` + transcript + `","cwd":"/w"}`

	require.NoError(t, runHook(t, paths, payload, "claude-code"))

	turns, err := capture.NewSpool(paths.SpoolPath).Read()
	require.NoError(t, err)
	require.Len(t, turns, 1)
	assert.Equal(t, "a1", turns[0].ID)
	assert.Equal(t, 9000, turns[0].Usage.CacheRead)
	assert.Equal(t, "feat/42-x", turns[0].Branch)
}

func TestHook_OpenCode_SpoolsTurns(t *testing.T) {
	paths := hookPaths(t)

	payload := `{"session_id":"s9","branch":"main","messages":[
		{"id":"m1","role":"assistant","modelID":"gpt-5.4","path":{"cwd":"/w"},
		 "tokens":{"input":10,"output":20,"reasoning":5,"cache":{"write":0,"read":300}},
		 "time":{"created":1784155308239}}]}`

	require.NoError(t, runHook(t, paths, payload, "opencode"))

	turns, err := capture.NewSpool(paths.SpoolPath).Read()
	require.NoError(t, err)
	require.Len(t, turns, 1)
	assert.Equal(t, 5, turns[0].Usage.Reasoning)
}

// An unparseable payload must be preserved, not dropped: a format change should
// cost an investigation, not the measurements.
func TestHook_UnparseablePayloadIsPreserved(t *testing.T) {
	paths := hookPaths(t)

	err := runHook(t, paths, "{not json", "claude-code")
	require.Error(t, err)

	data, readErr := os.ReadFile(paths.SpoolPath + ".raw")
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "{not json")
}

func TestHook_UnknownAgent(t *testing.T) {
	err := runHook(t, hookPaths(t), "{}", "some-other-agent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported agent")
}

func TestHook_MissingAgent(t *testing.T) {
	err := runHook(t, hookPaths(t), "{}")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent name is required")
}

func reconcileFixture(t *testing.T) (*store.SQLiteStore, config.Paths) {
	t.Helper()

	dir := t.TempDir()
	loader, err := pricing.NewLoader("")
	require.NoError(t, err)

	s, err := store.NewSQLiteStore(filepath.Join(dir, "t.db"), loader)
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	return s, config.Paths{
		BindingsPath: filepath.Join(dir, "bindings.json"),
		SpoolPath:    filepath.Join(dir, "spool.jsonl"),
	}
}

func runReconcile(t *testing.T, s store.Store, paths config.Paths) int {
	t.Helper()

	var recorded int

	app := &cli.App{
		Writer: &bytes.Buffer{},
		Action: func(c *cli.Context) error {
			recorded = reconcileSpool(c, s, paths)

			return nil
		},
	}
	require.NoError(t, app.RunContext(context.Background(), []string{"tok"}))

	return recorded
}

// The turn is attributed from its branch name, without any network call.
func TestReconcile_AttributesFromBranch(t *testing.T) {
	s, paths := reconcileFixture(t)

	require.NoError(t, capture.NewSpool(paths.SpoolPath).Append([]capture.Turn{{
		ID: "a1", Agent: "claude-code", Model: "claude-opus-5", SessionID: "s1",
		Branch: "feat/42-thing", Cwd: "/nonexistent",
		Usage: usage.Usage{InputFresh: 10, CacheRead: 900, Output: 5},
		At:    time.Now().UTC(),
	}}))

	bindings := attribution.NewStore(paths.BindingsPath)
	require.NoError(t, bindings.Bind("s1", "", attribution.Binding{Repo: "o/r", IssueNum: 42}))

	assert.Equal(t, 1, runReconcile(t, s, paths))

	entries, err := s.ListEntries(context.Background(), usage.Filter{Repo: "o/r"})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.NotNil(t, entries[0].IssueNum)
	assert.Equal(t, 42, *entries[0].IssueNum)
	assert.Equal(t, usage.SourceMeasured, entries[0].Source, "captured turns are measured, never estimated")
	assert.Equal(t, 900, entries[0].Usage.CacheRead)
	assert.Equal(t, "feat/42-thing", entries[0].Branch,
		"the branch is stored, not merely used to infer the issue and dropped")
}

// Draining twice must not double-count: storage is idempotent on the turn id.
func TestReconcile_IsIdempotent(t *testing.T) {
	s, paths := reconcileFixture(t)

	bindings := attribution.NewStore(paths.BindingsPath)
	require.NoError(t, bindings.Bind("s1", "", attribution.Binding{Repo: "o/r", IssueNum: 1}))

	turn := capture.Turn{
		ID: "a1", Agent: "claude-code", Model: "m", SessionID: "s1",
		Usage: usage.Usage{Output: 10}, At: time.Now().UTC(),
	}

	spool := capture.NewSpool(paths.SpoolPath)
	require.NoError(t, spool.Append([]capture.Turn{turn}))
	runReconcile(t, s, paths)

	// Replay the same turn, as a re-read transcript would produce.
	require.NoError(t, spool.Append([]capture.Turn{turn}))
	runReconcile(t, s, paths)

	entries, err := s.ListEntries(context.Background(), usage.Filter{Repo: "o/r"})
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

// Capture never depends on knowing the issue.
func TestReconcile_RecordsUnattributed(t *testing.T) {
	s, paths := reconcileFixture(t)

	bindings := attribution.NewStore(paths.BindingsPath)
	require.NoError(t, bindings.Bind("", "/w/repo", attribution.Binding{Repo: "o/r", IssueNum: 0}))

	require.NoError(t, capture.NewSpool(paths.SpoolPath).Append([]capture.Turn{{
		ID: "a1", Agent: "claude-code", Model: "m", SessionID: "sX",
		Cwd: "/w/repo", Branch: "main",
		Usage: usage.Usage{Output: 42}, At: time.Now().UTC(),
	}}))

	assert.Equal(t, 1, runReconcile(t, s, paths))

	groups, err := s.ListUnattributed(context.Background(), "o/r")
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, 42, groups[0].Usage.Output)
}

// A turn whose working directory is not a git checkout can never gain a
// repository. Leaving it spooled would stall every later turn behind it and
// grow the file without bound, so it is quarantined and the spool drains.
func TestReconcile_UnattributableTurnDoesNotStallTheSpool(t *testing.T) {
	s, paths := reconcileFixture(t)
	spool := capture.NewSpool(paths.SpoolPath)

	bindings := attribution.NewStore(paths.BindingsPath)
	require.NoError(t, bindings.Bind("good", "", attribution.Binding{Repo: "o/r", IssueNum: 1}))

	require.NoError(t, spool.Append([]capture.Turn{
		{ID: "bad", Agent: "claude-code", Model: "m", SessionID: "nowhere",
			Cwd:   filepath.Join(t.TempDir(), "not-a-repo"),
			Usage: usage.Usage{Output: 5}, At: time.Now().UTC()},
		{ID: "good1", Agent: "claude-code", Model: "m", SessionID: "good",
			Usage: usage.Usage{Output: 10}, At: time.Now().UTC()},
	}))

	assert.Equal(t, 1, runReconcile(t, s, paths), "the usable turn is recorded")
	assert.Equal(t, 0, spool.Pending(), "the spool drains rather than stalling")

	raw, err := os.ReadFile(paths.SpoolPath + ".raw")
	require.NoError(t, err)
	assert.Contains(t, string(raw), "bad", "the unusable turn is preserved, not discarded")
}

// The spec says --issue is optional. It was required, which made the synced
// spec claim a behaviour the code did not have.
func TestLog_WithoutIssue_RecordsUnattributed(t *testing.T) {
	s, _ := reconcileFixture(t)

	var buf bytes.Buffer

	app := &cli.App{
		Writer:   &buf,
		Commands: []*cli.Command{logCommand(s, nil)},
	}

	err := app.RunContext(context.Background(), []string{
		"tok", "log", "--repo", "owner/repo", "--agent", "claude-code",
		"--model", "claude-opus-5", "--input", "100", "--output", "50",
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "unattributed")

	entries, err := s.ListEntries(context.Background(), usage.Filter{Repo: "owner/repo"})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Nil(t, entries[0].IssueNum)
	assert.Equal(t, usage.SourceEstimated, entries[0].Source)
}
