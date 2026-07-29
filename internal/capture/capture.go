// Package capture reads token usage from the records agents keep of what the
// provider reported, rather than asking a model to estimate its own usage.
//
// Nothing here trusts a model's arithmetic: every count originates from the
// provider's response as the agent recorded it.
package capture

import (
	"time"

	"github.com/cdimonaco/tokenpile/internal/usage"
)

// Turn is one captured unit of work, normalized across agents.
//
// ID is the agent's own identifier for the message. It becomes the entry's
// primary key, which makes capture idempotent for free: re-reading a transcript
// or draining a spool twice cannot create duplicates.
type Turn struct {
	ID        string      `json:"id"`
	Agent     string      `json:"agent"`
	Model     string      `json:"model"`
	SessionID string      `json:"session_id,omitempty"`
	Cwd       string      `json:"cwd,omitempty"`
	Branch    string      `json:"branch,omitempty"`
	Usage     usage.Usage `json:"usage"`
	At        time.Time   `json:"at"`
}

// Valid reports whether a turn carries enough to be worth recording. A turn
// with no tokens is an agent event that cost nothing, not a measurement.
func (t Turn) Valid() bool {
	return t.ID != "" && t.Usage.TotalTokens() > 0
}
