package export_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/export"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

func newTestKey(t *testing.T) (ed25519.PrivateKey, ed25519.PublicKey) {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	return priv, pub
}

func testEntries() []usage.Entry {
	return []usage.Entry{
		{
			ID:        "1",
			Repo:      "o/r",
			IssueNum:  42,
			Agent:     "claude-code",
			Model:     "claude-sonnet-4-6",
			TokensIn:  1000,
			TokensOut: 500,
			At:        time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:        "2",
			Repo:      "o/r",
			IssueNum:  42,
			Agent:     "opencode",
			Model:     "gpt-4o",
			TokensIn:  2000,
			TokensOut: 800,
			At:        time.Date(2026, 7, 1, 11, 0, 0, 0, time.UTC),
		},
	}
}

func TestBuild_ProducesValidDocument(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, nil, priv, "0.1.0")
	require.NoError(t, err)

	assert.Equal(t, "3.0", doc.SchemaVersion)
	assert.NotEmpty(t, doc.Signature)
	assert.NotEmpty(t, doc.PublicKey)
	assert.Len(t, doc.Entries, 2)
	assert.Equal(t, "tokenpile/0.1.0", doc.ExportedBy)
}

func TestVerify_ValidDocument(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, nil, priv, "0.1.0")
	require.NoError(t, err)

	res, err := export.Verify(doc)
	require.NoError(t, err)
	assert.False(t, res.Legacy)
	assert.Equal(t, "3.0", res.SchemaVersion)
}

func TestVerify_TamperedEntries(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, nil, priv, "0.1.0")
	require.NoError(t, err)

	doc.Entries[0].TokensIn = 9999

	_, err = export.Verify(doc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tampered")
}

func TestVerify_InvalidSignature(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, nil, priv, "0.1.0")
	require.NoError(t, err)

	doc.Signature = "aW52YWxpZA=="

	_, err = export.Verify(doc)
	assert.Error(t, err)
}

func TestVerify_SurvivesJSONRoundtrip(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, []export.IssueBudget{
		{Repo: "o/r", IssueNum: 42, Amount: 12.5},
	}, priv, "0.1.0")
	require.NoError(t, err)

	data, err := json.Marshal(doc)
	require.NoError(t, err)

	var decoded export.Document
	require.NoError(t, json.Unmarshal(data, &decoded))

	_, err = export.Verify(&decoded)
	assert.NoError(t, err)
}

func TestBuild_EmptyEntries(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(nil, nil, nil, priv, "1.0.0")
	require.NoError(t, err)
	assert.Empty(t, doc.Entries)

	_, err = export.Verify(doc)
	assert.NoError(t, err)
}

func TestBuild_SignatureChangesWhenEntriesChange(t *testing.T) {
	priv, _ := newTestKey(t)

	doc1, err := export.Build(testEntries(), nil, nil, priv, "0.1.0")
	require.NoError(t, err)

	modified := testEntries()
	modified[0].TokensIn = 99999

	doc2, err := export.Build(modified, nil, nil, priv, "0.1.0")
	require.NoError(t, err)

	assert.NotEqual(t, doc1.Signature, doc2.Signature)
}

func TestVerify_WrongKey(t *testing.T) {
	priv1, _ := newTestKey(t)
	_, pub2 := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, nil, priv1, "0.1.0")
	require.NoError(t, err)

	doc.PublicKey = base64.StdEncoding.EncodeToString(pub2)

	_, err = export.Verify(doc)
	assert.Error(t, err)
}

func TestVerify_TruncatedPublicKey(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, nil, priv, "0.1.0")
	require.NoError(t, err)

	doc.PublicKey = base64.StdEncoding.EncodeToString([]byte("tooshort"))

	_, err = export.Verify(doc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key size")
}

func TestBuild_SchemaVersionIsV3(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, nil, priv, "0.2.0")
	require.NoError(t, err)
	assert.Equal(t, "3.0", doc.SchemaVersion)
}

func TestBuild_SessionsIncludedInDocument(t *testing.T) {
	priv, _ := newTestKey(t)

	now := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	end := now.Add(30 * time.Minute)
	sessions := []usage.Session{
		{
			ID:        "s1",
			Repo:      "o/r",
			IssueNum:  42,
			StartedAt: now,
			EndedAt:   &end,
			Note:      "initial spike",
			Tags:      []string{"spike", "debug"},
		},
		{
			ID:        "s2",
			Repo:      "o/r",
			IssueNum:  42,
			StartedAt: now.Add(time.Hour),
		},
	}

	doc, err := export.Build(testEntries(), sessions, nil, priv, "0.2.0")
	require.NoError(t, err)
	require.Len(t, doc.Sessions, 2)

	assert.Equal(t, "s1", doc.Sessions[0].ID)
	assert.Equal(t, "initial spike", doc.Sessions[0].Note)
	assert.Equal(t, []string{"spike", "debug"}, doc.Sessions[0].Tags)
	assert.NotEmpty(t, doc.Sessions[0].EndedAt)

	assert.Equal(t, "s2", doc.Sessions[1].ID)
	assert.Empty(t, doc.Sessions[1].EndedAt)
}

func TestBuild_SessionsNilProducesEmptyBlock(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, nil, priv, "0.2.0")
	require.NoError(t, err)
	assert.Empty(t, doc.Sessions)
}

func TestBuild_BudgetsIncludedInDocument(t *testing.T) {
	priv, _ := newTestKey(t)

	budgets := []export.IssueBudget{
		{Repo: "o/r", IssueNum: 42, Amount: 50.0},
	}

	doc, err := export.Build(testEntries(), nil, budgets, priv, "0.2.0")
	require.NoError(t, err)
	require.Len(t, doc.Budgets, 1)
	assert.Equal(t, "o/r", doc.Budgets[0].Repo)
	assert.Equal(t, 42, doc.Budgets[0].IssueNum)
	assert.InEpsilon(t, 50.0, doc.Budgets[0].Amount, 0.001)
}

func TestBuild_BudgetsNilProducesEmptyBlock(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, nil, priv, "0.2.0")
	require.NoError(t, err)
	assert.Empty(t, doc.Budgets)
}

func TestVerify_TamperedSessions(t *testing.T) {
	priv, _ := newTestKey(t)

	now := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	sessions := []usage.Session{{ID: "s1", Repo: "o/r", IssueNum: 42, StartedAt: now}}

	doc, err := export.Build(testEntries(), sessions, nil, priv, "0.2.0")
	require.NoError(t, err)

	doc.Sessions[0].Note = "tampered"

	_, err = export.Verify(doc)
	assert.Error(t, err, "tampering sessions must invalidate the signature")
}

func TestVerify_TamperedBudgets(t *testing.T) {
	priv, _ := newTestKey(t)

	budgets := []export.IssueBudget{{Repo: "o/r", IssueNum: 42, Amount: 50.0}}

	doc, err := export.Build(testEntries(), nil, budgets, priv, "0.2.0")
	require.NoError(t, err)

	doc.Budgets[0].Amount = 9999.0

	_, err = export.Verify(doc)
	assert.Error(t, err, "tampering budgets must invalidate the signature")
}

func TestVerify_TamperedExportedAt(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, nil, priv, "0.2.0")
	require.NoError(t, err)

	doc.ExportedAt = "1999-01-01T00:00:00Z"

	_, err = export.Verify(doc)
	assert.Error(t, err, "tampering exported_at must invalidate the signature")
}

func TestVerify_UnsupportedSchemaVersion(t *testing.T) {
	priv, _ := newTestKey(t)

	doc, err := export.Build(testEntries(), nil, nil, priv, "0.2.0")
	require.NoError(t, err)

	doc.SchemaVersion = "9.9"

	_, err = export.Verify(doc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported schema version")
}

func TestVerify_LegacyV2Fixture(t *testing.T) {
	data, err := os.ReadFile("testdata/export_v2.json")
	require.NoError(t, err)

	var doc export.Document
	require.NoError(t, json.Unmarshal(data, &doc))

	res, err := export.Verify(&doc)
	require.NoError(t, err)
	assert.True(t, res.Legacy)
	assert.Equal(t, "2.0", res.SchemaVersion)
}

func TestVerify_LegacyV2Fixture_TamperedEntriesFail(t *testing.T) {
	data, err := os.ReadFile("testdata/export_v2.json")
	require.NoError(t, err)

	var doc export.Document
	require.NoError(t, json.Unmarshal(data, &doc))

	doc.Entries[0].TokensIn = 424242

	_, err = export.Verify(&doc)
	assert.Error(t, err)
}
