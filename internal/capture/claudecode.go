package capture

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cdimonaco/tokenpile/internal/usage"
)

// AgentClaudeCode is the agent name recorded for turns read from a Claude Code
// transcript.
const AgentClaudeCode = "claude-code"

// HookPayload is the JSON Claude Code writes to a hook's stdin. Only the fields
// tokenpile needs are declared; the payload carries more.
//
// Note what is absent: no token counts, and no model outside SessionStart. The
// hook says when a turn ended and where the transcript is; the numbers come
// from the transcript.
type HookPayload struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
}

// transcriptLine is one JSONL record. Assistant records carry the usage the
// provider reported for that message.
type transcriptLine struct {
	Type      string `json:"type"`
	UUID      string `json:"uuid"`
	SessionID string `json:"sessionId"`
	Cwd       string `json:"cwd"`
	GitBranch string `json:"gitBranch"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		Role  string `json:"role"`
		Model string `json:"model"`
		Usage *struct {
			InputTokens              int `json:"input_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
			OutputTokens             int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// ErrNoUsage reports that a transcript held no assistant message with usage.
var ErrNoUsage = errors.New("transcript contains no usage")

// ReadClaudeCodeTranscript extracts every assistant turn carrying usage.
//
// The transcript accumulates over a session, so this returns all turns rather
// than only the newest: deduplication happens on the entry's primary key, which
// is cheaper and more reliable than tracking a read watermark per session.
//
// Anthropic bills thinking tokens inside output and reports no separate
// reasoning count, so Reasoning stays zero — correct, not missing.
func ReadClaudeCodeTranscript(r io.Reader) ([]Turn, error) {
	scanner := bufio.NewScanner(r)
	// Transcript lines carry whole tool outputs and routinely exceed the
	// default 64KB limit.
	scanner.Buffer(make([]byte, 0, 1024*1024), 16*1024*1024)

	var (
		turns   []Turn
		badLine int
	)

	for scanner.Scan() {
		var line transcriptLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			// One unparseable line must not discard a whole session's worth of
			// measurements; count it so the caller can report it.
			badLine++

			continue
		}

		if line.Message.Role != "assistant" || line.Message.Usage == nil {
			continue
		}

		at, err := time.Parse(time.RFC3339, line.Timestamp)
		if err != nil {
			at = time.Now().UTC()
		}

		turn := Turn{
			ID:        line.UUID,
			Agent:     AgentClaudeCode,
			Model:     line.Message.Model,
			SessionID: line.SessionID,
			Cwd:       line.Cwd,
			Branch:    line.GitBranch,
			At:        at.UTC(),
			Usage: usage.Usage{
				InputFresh: line.Message.Usage.InputTokens,
				CacheWrite: line.Message.Usage.CacheCreationInputTokens,
				CacheRead:  line.Message.Usage.CacheReadInputTokens,
				Output:     line.Message.Usage.OutputTokens,
			},
		}

		if turn.Valid() {
			turns = append(turns, turn)
		}
	}

	if err := scanner.Err(); err != nil {
		return turns, fmt.Errorf("read transcript: %w", err)
	}

	if badLine > 0 && len(turns) == 0 {
		return nil, fmt.Errorf("transcript unparseable: %d malformed lines", badLine)
	}

	if len(turns) == 0 {
		return nil, ErrNoUsage
	}

	return turns, nil
}

// ReadClaudeCodeTranscriptFile is ReadClaudeCodeTranscript over a path.
func ReadClaudeCodeTranscriptFile(path string) ([]Turn, error) {
	f, err := os.Open(path) // #nosec G304 -- path comes from the agent's own hook payload
	if err != nil {
		return nil, fmt.Errorf("open transcript: %w", err)
	}
	defer func() { _ = f.Close() }()

	return ReadClaudeCodeTranscript(f)
}
