package capture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/capture"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

// Two real assistant records as Claude Code writes them, trimmed to the fields
// tokenpile reads, plus a user record that carries no usage.
const claudeTranscript = `
{"type":"user","uuid":"u1","sessionId":"s1","message":{"role":"user","content":"hi"}}
{"type":"assistant","uuid":"a1","sessionId":"s1","cwd":"/w/repo","gitBranch":"feat/42-thing","timestamp":"2026-07-29T10:37:52.695Z","message":{"role":"assistant","model":"claude-opus-5","usage":{"input_tokens":2,"cache_creation_input_tokens":18110,"cache_read_input_tokens":12725,"output_tokens":255}}}
{"type":"assistant","uuid":"a2","sessionId":"s1","cwd":"/w/repo","gitBranch":"feat/42-thing","timestamp":"2026-07-29T10:39:00.000Z","message":{"role":"assistant","model":"claude-opus-5","usage":{"input_tokens":1,"cache_creation_input_tokens":0,"cache_read_input_tokens":31000,"output_tokens":40}}}
`

func TestReadClaudeCodeTranscript(t *testing.T) {
	turns, err := capture.ReadClaudeCodeTranscript(strings.NewReader(claudeTranscript))

	require.NoError(t, err)
	require.Len(t, turns, 2, "user records carry no usage and are skipped")

	first := turns[0]
	assert.Equal(t, "a1", first.ID)
	assert.Equal(t, capture.AgentClaudeCode, first.Agent)
	assert.Equal(t, "claude-opus-5", first.Model)
	assert.Equal(t, "s1", first.SessionID)
	assert.Equal(t, "/w/repo", first.Cwd)
	assert.Equal(t, "feat/42-thing", first.Branch)
	assert.Equal(t, usage.Usage{
		InputFresh: 2, CacheWrite: 18110, CacheRead: 12725, Output: 255,
	}, first.Usage)
}

// Anthropic bills thinking inside output and reports no separate count, so zero
// is the correct value rather than a gap.
func TestReadClaudeCodeTranscript_ReasoningIsZero(t *testing.T) {
	turns, err := capture.ReadClaudeCodeTranscript(strings.NewReader(claudeTranscript))

	require.NoError(t, err)

	for _, turn := range turns {
		assert.Equal(t, 0, turn.Usage.Reasoning)
	}
}

func TestReadClaudeCodeTranscript_NoUsage(t *testing.T) {
	_, err := capture.ReadClaudeCodeTranscript(strings.NewReader(
		`{"type":"user","uuid":"u1","message":{"role":"user"}}` + "\n"))

	require.ErrorIs(t, err, capture.ErrNoUsage)
}

// One corrupt line must not cost the whole session's measurements.
func TestReadClaudeCodeTranscript_SkipsMalformedLine(t *testing.T) {
	input := "{ this is not json\n" + claudeTranscript

	turns, err := capture.ReadClaudeCodeTranscript(strings.NewReader(input))

	require.NoError(t, err)
	assert.Len(t, turns, 2)
}

func TestReadClaudeCodeTranscript_AllMalformedIsAnError(t *testing.T) {
	_, err := capture.ReadClaudeCodeTranscript(strings.NewReader("nonsense\nmore nonsense\n"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unparseable")
}

const openCodePayload = `{
  "session_id": "sess-9",
  "branch": "fix/7-bug",
  "messages": [
    {"id":"m1","role":"assistant","modelID":"deepseek-v4-flash-free",
     "cost":0.42,"path":{"cwd":"/w/other"},
     "tokens":{"input":813,"output":24,"reasoning":10,"cache":{"write":0,"read":107136}},
     "time":{"created":1784155308239}},
    {"id":"m2","role":"user","tokens":{"input":0,"output":0,"cache":{"write":0,"read":0}}}
  ]
}`

func TestReadOpenCodePayload(t *testing.T) {
	turns, err := capture.ReadOpenCodePayload(strings.NewReader(openCodePayload))

	require.NoError(t, err)
	require.Len(t, turns, 1)

	turn := turns[0]
	assert.Equal(t, "m1", turn.ID)
	assert.Equal(t, capture.AgentOpenCode, turn.Agent)
	assert.Equal(t, "deepseek-v4-flash-free", turn.Model)
	assert.Equal(t, "sess-9", turn.SessionID)
	assert.Equal(t, "fix/7-bug", turn.Branch)
	assert.Equal(t, usage.Usage{
		InputFresh: 813, CacheWrite: 0, CacheRead: 107136, Output: 24, Reasoning: 10,
	}, turn.Usage)
}

// opencode computes its own cost. tokenpile recomputes from its own pricing so
// figures stay comparable across agents, so that field must not leak in.
func TestReadOpenCodePayload_IgnoresAgentCost(t *testing.T) {
	turns, err := capture.ReadOpenCodePayload(strings.NewReader(openCodePayload))

	require.NoError(t, err)
	require.Len(t, turns, 1)
	assert.Equal(t, 10, turns[0].Usage.Reasoning, "reasoning is carried, not folded into output")
	assert.Equal(t, 24, turns[0].Usage.Output, "reasoning is not added on top of output")
}

func TestSpool_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "spool", "capture.jsonl")
	s := capture.NewSpool(path)

	turns := []capture.Turn{
		{ID: "a1", Agent: "claude-code", Usage: usage.Usage{InputFresh: 10}},
		{ID: "a2", Agent: "claude-code", Usage: usage.Usage{Output: 20}},
	}

	require.NoError(t, s.Append(turns))
	assert.Equal(t, 2, s.Pending())

	got, err := s.Read()
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "a1", got[0].ID)

	require.NoError(t, s.Clear())
	assert.Equal(t, 0, s.Pending())
}

func TestSpool_MissingFileIsEmpty(t *testing.T) {
	s := capture.NewSpool(filepath.Join(t.TempDir(), "nothing.jsonl"))

	got, err := s.Read()

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestSpool_AppendIsAdditive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture.jsonl")
	s := capture.NewSpool(path)

	require.NoError(t, s.Append([]capture.Turn{{ID: "a1", Usage: usage.Usage{Output: 1}}}))
	require.NoError(t, s.Append([]capture.Turn{{ID: "a2", Usage: usage.Usage{Output: 1}}}))

	assert.Equal(t, 2, s.Pending())
}

func TestSpool_RawPreservesUnparseablePayload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture.jsonl")
	s := capture.NewSpool(path)

	require.NoError(t, s.AppendRaw("claude-code", []byte("{broken"), capture.ErrNoUsage))

	data, err := readFile(path + ".raw")
	require.NoError(t, err)
	assert.Contains(t, data, "{broken")
	assert.Contains(t, data, "claude-code")
}

func TestTurn_ValidRequiresTokens(t *testing.T) {
	assert.False(t, capture.Turn{ID: "a1"}.Valid(), "a turn that cost nothing is not a measurement")
	assert.False(t, capture.Turn{Usage: usage.Usage{Output: 1}}.Valid(), "an id is required to deduplicate")
	assert.True(t, capture.Turn{ID: "a1", Usage: usage.Usage{Output: 1}}.Valid())
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)

	return string(data), err
}
