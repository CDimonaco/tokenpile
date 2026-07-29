package export

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCanonicalJSON_PinnedBytes pins the exact canonical serialization the
// signature is computed over. If a Go upgrade changes JSON formatting
// (notably float rendering), this fails loudly instead of silently breaking
// verification of previously signed exports.
func TestCanonicalJSON_PinnedBytes(t *testing.T) {
	doc := Document{
		SchemaVersion: "4.0",
		ExportedAt:    "2026-07-01T12:00:00Z",
		ExportedBy:    "tokenpile/test",
		PublicKey:     "IVL40Zt5HSRFMkLhXy6rbLfP+ntqXtMAl5YOBpiB2xI=",
		Entries: []entryJSON{
			{
				ID: "e1", Repo: "o/r", IssueNum: 42, Agent: "claude-code",
				Model: "claude-sonnet-4-6", InputFresh: 1000, Output: 500, Source: "estimated",
				At: "2026-07-01T10:00:00Z",
			},
		},
		Budgets: []budgetJSON{{Repo: "o/r", IssueNum: 42, Amount: 3.14159}},
	}

	got, err := canonicalJSON(doc)
	require.NoError(t, err)

	//nolint:lll
	want := `{"budgets":[{"amount":3.14159,"issue_num":42,"repo":"o/r"}],"entries":[{"agent":"claude-code","at":"2026-07-01T10:00:00Z","cache_read":0,"cache_write":0,"id":"e1","input_fresh":1000,"issue_num":42,"model":"claude-sonnet-4-6","output":500,"reasoning":0,"repo":"o/r","source":"estimated"}],"exported_at":"2026-07-01T12:00:00Z","exported_by":"tokenpile/test","public_key":"IVL40Zt5HSRFMkLhXy6rbLfP+ntqXtMAl5YOBpiB2xI=","schema_version":"4.0","signature":""}`

	assert.Equal(t, want, string(got))
}
