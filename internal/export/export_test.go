package export_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/cdimonaco/tokenpile/internal/domain"
	"github.com/cdimonaco/tokenpile/internal/export"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestKey(t *testing.T) (ed25519.PrivateKey, ed25519.PublicKey) {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	return priv, pub
}

func testEntries() []domain.UsageEntry {
	return []domain.UsageEntry{
		{ID: "1", Repo: "o/r", IssueNum: 42, Agent: "claude-code", Model: "claude-sonnet-4-6", TokensIn: 1000, TokensOut: 500, At: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)},
		{ID: "2", Repo: "o/r", IssueNum: 42, Agent: "opencode", Model: "gpt-4o", TokensIn: 2000, TokensOut: 800, At: time.Date(2026, 7, 1, 11, 0, 0, 0, time.UTC)},
	}
}

func TestBuild_ProducesValidDocument(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), priv, "0.1.0")
	require.NoError(t, err)

	assert.Equal(t, "1.0", doc.SchemaVersion)
	assert.NotEmpty(t, doc.Signature)
	assert.NotEmpty(t, doc.PublicKey)
	assert.Len(t, doc.Entries, 2)
	assert.Equal(t, "tokenpile/0.1.0", doc.ExportedBy)
}

func TestVerify_ValidDocument(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), priv, "0.1.0")
	require.NoError(t, err)

	err = export.Verify(doc)
	assert.NoError(t, err)
}

func TestVerify_TamperedEntries(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), priv, "0.1.0")
	require.NoError(t, err)

	doc.Entries[0].TokensIn = 9999

	err = export.Verify(doc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tampered")
}

func TestVerify_InvalidSignature(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), priv, "0.1.0")
	require.NoError(t, err)

	doc.Signature = "aW52YWxpZA=="

	err = export.Verify(doc)
	assert.Error(t, err)
}

func TestCanonicalJSON_DeterministicAcrossBuilds(t *testing.T) {
	priv, _ := newTestKey(t)
	entries := testEntries()

	doc1, err := export.Build(entries, priv, "0.1.0")
	require.NoError(t, err)

	doc2, err := export.Build(entries, priv, "0.1.0")
	require.NoError(t, err)

	assert.Equal(t, doc1.Signature, doc2.Signature)
}
