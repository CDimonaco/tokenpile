package export

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/cdimonaco/tokenpile/internal/schema"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

var Schema = schema.ExportSchema

const SchemaVersion = "1.0"

type entryJSON struct {
	ID        string `json:"id"`
	Repo      string `json:"repo"`
	IssueNum  int    `json:"issue_num"`
	Agent     string `json:"agent"`
	Model     string `json:"model"`
	TokensIn  int    `json:"tokens_in"`
	TokensOut int    `json:"tokens_out"`
	SessionID string `json:"session_id,omitempty"`
	At        string `json:"at"`
}

type Document struct {
	SchemaVersion string      `json:"schema_version"`
	ExportedAt    string      `json:"exported_at"`
	ExportedBy    string      `json:"exported_by"`
	PublicKey     string      `json:"public_key"`
	Entries       []entryJSON `json:"entries"`
	Signature     string      `json:"signature"`
}

func Build(entries []usage.Entry, priv ed25519.PrivateKey, version string) (*Document, error) {
	jsonEntries := make([]entryJSON, len(entries))
	for i, e := range entries {
		jsonEntries[i] = entryJSON{
			ID:        e.ID,
			Repo:      e.Repo,
			IssueNum:  e.IssueNum,
			Agent:     e.Agent,
			Model:     e.Model,
			TokensIn:  e.TokensIn,
			TokensOut: e.TokensOut,
			SessionID: e.SessionID,
			At:        e.At.UTC().Format(time.RFC3339),
		}
	}

	canonical, err := canonicalJSON(jsonEntries)
	if err != nil {
		return nil, fmt.Errorf("canonical JSON: %w", err)
	}

	digest := sha256.Sum256(canonical)
	sig := ed25519.Sign(priv, digest[:])

	pub := priv.Public().(ed25519.PublicKey)

	return &Document{
		SchemaVersion: SchemaVersion,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		ExportedBy:    "tokenpile/" + version,
		PublicKey:     base64.StdEncoding.EncodeToString(pub),
		Entries:       jsonEntries,
		Signature:     base64.StdEncoding.EncodeToString(sig),
	}, nil
}

func Verify(doc *Document) error {
	pubBytes, err := base64.StdEncoding.DecodeString(doc.PublicKey)
	if err != nil {
		return fmt.Errorf("decode public key: %w", err)
	}

	if len(pubBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size: got %d, want %d", len(pubBytes), ed25519.PublicKeySize)
	}

	pub := ed25519.PublicKey(pubBytes)

	sigBytes, err := base64.StdEncoding.DecodeString(doc.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}

	canonical, err := canonicalJSON(doc.Entries)
	if err != nil {
		return fmt.Errorf("canonical JSON: %w", err)
	}

	digest := sha256.Sum256(canonical)

	if !ed25519.Verify(pub, digest[:], sigBytes) {
		return fmt.Errorf("signature invalid: entries have been tampered with")
	}

	return nil
}

func canonicalJSON(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var normalized any
	if err = json.Unmarshal(raw, &normalized); err != nil {
		return nil, err
	}

	return marshalCanonical(normalized)
}

func marshalCanonical(v any) ([]byte, error) {
	switch val := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		buf := []byte{'{'}

		for i, k := range keys {
			if i > 0 {
				buf = append(buf, ',')
			}

			keyBytes, err := json.Marshal(k)
			if err != nil {
				return nil, err
			}

			buf = append(buf, keyBytes...)
			buf = append(buf, ':')

			valBytes, err := marshalCanonical(val[k])
			if err != nil {
				return nil, err
			}

			buf = append(buf, valBytes...)
		}

		buf = append(buf, '}')

		return buf, nil

	case []any:
		buf := []byte{'['}

		for i, item := range val {
			if i > 0 {
				buf = append(buf, ',')
			}

			itemBytes, err := marshalCanonical(item)
			if err != nil {
				return nil, err
			}

			buf = append(buf, itemBytes...)
		}

		buf = append(buf, ']')

		return buf, nil

	default:
		return json.Marshal(val)
	}
}
