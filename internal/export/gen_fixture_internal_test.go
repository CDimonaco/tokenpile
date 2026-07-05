package export

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestGenerateFixtures(t *testing.T) {
	if os.Getenv("GEN_FIXTURES") == "" {
		t.Skip("generator")
	}

	seed := bytes.Repeat([]byte{0x42}, ed25519.SeedSize)
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)

	entries := []entryJSON{
		{
			ID: "e1", Repo: "o/r", IssueNum: 42, Agent: "claude-code",
			Model: "claude-sonnet-4-6", TokensIn: 1000, TokensOut: 500,
			At: "2026-07-01T10:00:00Z",
		},
	}

	canonical, err := canonicalJSON(entries)
	if err != nil {
		t.Fatal(err)
	}

	digest := sha256.Sum256(canonical)
	sig := ed25519.Sign(priv, digest[:])

	doc := Document{
		SchemaVersion: "2.0",
		ExportedAt:    "2026-07-01T12:00:00Z",
		ExportedBy:    "tokenpile/0.9.0",
		PublicKey:     base64.StdEncoding.EncodeToString(pub),
		Entries:       entries,
		Sessions: []sessionJSON{
			{ID: "s1", Repo: "o/r", IssueNum: 42, StartedAt: "2026-07-01T10:00:00Z", Note: "legacy note"},
		},
		Budgets:   []budgetJSON{{Repo: "o/r", IssueNum: 42, Amount: 12.5}},
		Signature: base64.StdEncoding.EncodeToString(sig),
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if err = os.MkdirAll("testdata", 0o755); err != nil {
		t.Fatal(err)
	}

	if err = os.WriteFile("testdata/export_v2.json", append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	pinnedDoc := Document{
		SchemaVersion: "3.0",
		ExportedAt:    "2026-07-01T12:00:00Z",
		ExportedBy:    "tokenpile/test",
		PublicKey:     base64.StdEncoding.EncodeToString(pub),
		Entries:       entries,
		Budgets:       []budgetJSON{{Repo: "o/r", IssueNum: 42, Amount: 3.14159}},
	}

	pinned, err := canonicalJSON(pinnedDoc)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("PINNED_CANONICAL: %s\n", pinned)
}
