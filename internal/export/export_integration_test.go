package export_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/export"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

func TestExport_RoundTrip_EmptyEntries(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	doc, err := export.Build(nil, nil, nil, priv, "test")
	require.NoError(t, err)

	assert.Equal(t, export.SchemaVersion, doc.SchemaVersion)
	assert.Empty(t, doc.Entries)

	res, err := export.Verify(doc)
	require.NoError(t, err)
	assert.False(t, res.Legacy)
}

func TestExport_RoundTrip_WithEntries(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	entries := []usage.Entry{
		{
			ID:        "e1",
			Repo:      "owner/repo",
			IssueNum:  42,
			Agent:     "claude-code",
			Model:     "claude-sonnet-4-6",
			TokensIn:  1000,
			TokensOut: 500,
			SessionID: "s1",
			At:        time.Now().UTC().Truncate(time.Second),
		},
		{
			ID:        "e2",
			Repo:      "owner/repo",
			IssueNum:  43,
			Agent:     "opencode",
			Model:     "gpt-4o",
			TokensIn:  200,
			TokensOut: 100,
			At:        time.Now().UTC().Truncate(time.Second),
		},
	}

	doc, err := export.Build(entries, nil, nil, priv, "1.0.0")
	require.NoError(t, err)
	require.Len(t, doc.Entries, 2)

	_, err = export.Verify(doc)
	require.NoError(t, err)
}

func TestExport_Verify_TamperedEntries(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	entries := []usage.Entry{
		{ID: "e1", Repo: "o/r", IssueNum: 1, Agent: "a", Model: "m", TokensIn: 100, TokensOut: 50, At: time.Now()},
	}

	doc, err := export.Build(entries, nil, nil, priv, "test")
	require.NoError(t, err)

	doc.Entries[0].TokensIn = 99999

	_, err = export.Verify(doc)
	assert.Error(t, err)
}

func TestExport_Verify_InvalidPublicKey(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	doc, err := export.Build(nil, nil, nil, priv, "test")
	require.NoError(t, err)

	doc.PublicKey = "not!valid!base64!!!"

	_, err = export.Verify(doc)
	assert.Error(t, err)
}

func TestExport_Verify_WrongPublicKey(t *testing.T) {
	_, priv1, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pub2, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	entries := []usage.Entry{
		{ID: "e1", Repo: "o/r", IssueNum: 1, Agent: "a", Model: "m", TokensIn: 100, TokensOut: 50, At: time.Now()},
	}

	doc, err := export.Build(entries, nil, nil, priv1, "test")
	require.NoError(t, err)

	doc.PublicKey = base64.StdEncoding.EncodeToString(pub2)

	_, err = export.Verify(doc)
	assert.Error(t, err)
}

func TestExport_Verify_CorruptedSignature(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	doc, err := export.Build(nil, nil, nil, priv, "test")
	require.NoError(t, err)

	doc.Signature = base64.StdEncoding.EncodeToString(make([]byte, 64))

	_, err = export.Verify(doc)
	assert.Error(t, err)
}
