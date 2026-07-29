package capture

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/cdimonaco/tokenpile/internal/usage"
)

// AgentOpenCode is the agent name recorded for turns read from opencode.
const AgentOpenCode = "opencode"

// openCodeMessage mirrors the shape opencode stores per assistant message and
// exposes through its SDK. The plugin forwards these rather than tokenpile
// reading opencode's database directly: reaching into another application's
// private storage would break the moment it changes its schema.
type openCodeMessage struct {
	ID      string  `json:"id"`
	Role    string  `json:"role"`
	ModelID string  `json:"modelID"`
	Cost    float64 `json:"cost"`
	Path    struct {
		Cwd  string `json:"cwd"`
		Root string `json:"root"`
	} `json:"path"`
	Tokens struct {
		Input     int `json:"input"`
		Output    int `json:"output"`
		Reasoning int `json:"reasoning"`
		Cache     struct {
			Write int `json:"write"`
			Read  int `json:"read"`
		} `json:"cache"`
	} `json:"tokens"`
	Time struct {
		Created int64 `json:"created"`
	} `json:"time"`
}

// openCodePayload is what the plugin sends: the session it belongs to plus the
// messages produced in it.
type openCodePayload struct {
	SessionID string            `json:"session_id"`
	Branch    string            `json:"branch"`
	Messages  []openCodeMessage `json:"messages"`
}

// ReadOpenCodePayload converts the plugin's JSON into turns.
//
// opencode computes its own cost, which is deliberately ignored: cost is
// recomputed from tokenpile's pricing so figures stay comparable across agents.
//
// Unlike Claude Code, opencode reports reasoning tokens separately. They are a
// subset of output and are carried as such, never added to it.
func ReadOpenCodePayload(r io.Reader) ([]Turn, error) {
	var payload openCodePayload
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode opencode payload: %w", err)
	}

	var turns []Turn

	for _, m := range payload.Messages {
		if m.Role != "assistant" {
			continue
		}

		at := time.Now().UTC()
		if m.Time.Created > 0 {
			at = time.UnixMilli(m.Time.Created).UTC()
		}

		turn := Turn{
			ID:        m.ID,
			Agent:     AgentOpenCode,
			Model:     m.ModelID,
			SessionID: payload.SessionID,
			Cwd:       m.Path.Cwd,
			Branch:    payload.Branch,
			At:        at,
			Usage: usage.Usage{
				InputFresh: m.Tokens.Input,
				CacheWrite: m.Tokens.Cache.Write,
				CacheRead:  m.Tokens.Cache.Read,
				Output:     m.Tokens.Output,
				Reasoning:  m.Tokens.Reasoning,
			},
		}

		if turn.Valid() {
			turns = append(turns, turn)
		}
	}

	if len(turns) == 0 {
		return nil, ErrNoUsage
	}

	return turns, nil
}
