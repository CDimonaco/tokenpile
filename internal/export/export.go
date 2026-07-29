package export

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/cdimonaco/tokenpile/internal/schema"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

var Schema = schema.ExportSchema

const (
	// SchemaVersion is the version written by Build. The signature covers the
	// canonical JSON of the whole document with the signature field emptied.
	SchemaVersion = "4.0"
)

// VerifyResult reports how a document was verified.
type VerifyResult struct {
	SchemaVersion string
}

type entryJSON struct {
	ID          string   `json:"id"`
	Repo        string   `json:"repo"`
	IssueNum    int      `json:"issue_num"`
	IssueTitle  string   `json:"issue_title,omitempty"`
	IssueLabels []string `json:"issue_labels,omitempty"`
	Agent       string   `json:"agent"`
	Model       string   `json:"model"`
	InputFresh  int      `json:"input_fresh"`
	CacheWrite  int      `json:"cache_write"`
	CacheRead   int      `json:"cache_read"`
	Output      int      `json:"output"`
	Reasoning   int      `json:"reasoning"`
	Source      string   `json:"source"`
	SessionID   string   `json:"session_id,omitempty"`
	At          string   `json:"at"`
}

type sessionJSON struct {
	ID        string   `json:"id"`
	Repo      string   `json:"repo"`
	IssueNum  int      `json:"issue_num"`
	StartedAt string   `json:"started_at"`
	EndedAt   string   `json:"ended_at,omitempty"`
	Note      string   `json:"note,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

type budgetJSON struct {
	Repo     string  `json:"repo"`
	IssueNum int     `json:"issue_num"`
	Amount   float64 `json:"amount"`
}

type Document struct {
	SchemaVersion string        `json:"schema_version"`
	ExportedAt    string        `json:"exported_at"`
	ExportedBy    string        `json:"exported_by"`
	PublicKey     string        `json:"public_key"`
	Entries       []entryJSON   `json:"entries"`
	Sessions      []sessionJSON `json:"sessions,omitempty"`
	Budgets       []budgetJSON  `json:"budgets,omitempty"`
	Signature     string        `json:"signature"`
}

// IssueBudget carries a budget amount for a specific issue, used in exports.
type IssueBudget struct {
	Repo     string
	IssueNum int
	Amount   float64
}

func Build(
	entries []usage.Entry,
	sessions []usage.Session,
	budgets []IssueBudget,
	priv ed25519.PrivateKey,
	version string,
) (*Document, error) {
	jsonEntries := make([]entryJSON, len(entries))
	for i, e := range entries {
		jsonEntries[i] = entryJSON{
			ID:          e.ID,
			Repo:        e.Repo,
			IssueNum:    e.IssueNum,
			IssueTitle:  e.IssueTitle,
			IssueLabels: e.IssueLabels,
			Agent:       e.Agent,
			Model:       e.Model,
			InputFresh:  e.Usage.InputFresh,
			CacheWrite:  e.Usage.CacheWrite,
			CacheRead:   e.Usage.CacheRead,
			Output:      e.Usage.Output,
			Reasoning:   e.Usage.Reasoning,
			Source:      string(e.Source),
			SessionID:   e.SessionID,
			At:          e.At.UTC().Format(time.RFC3339),
		}
	}

	pub, ok := priv.Public().(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("private key is not ed25519")
	}

	jsonSessions := make([]sessionJSON, 0, len(sessions))
	for _, s := range sessions {
		sj := sessionJSON{
			ID:        s.ID,
			Repo:      s.Repo,
			IssueNum:  s.IssueNum,
			StartedAt: s.StartedAt.UTC().Format(time.RFC3339),
			Note:      s.Note,
			Tags:      s.Tags,
		}

		if s.EndedAt != nil {
			sj.EndedAt = s.EndedAt.UTC().Format(time.RFC3339)
		}

		jsonSessions = append(jsonSessions, sj)
	}

	jsonBudgets := make([]budgetJSON, 0, len(budgets))
	for _, b := range budgets {
		jsonBudgets = append(jsonBudgets, budgetJSON(b))
	}

	doc := &Document{
		SchemaVersion: SchemaVersion,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		ExportedBy:    "tokenpile/" + version,
		PublicKey:     base64.StdEncoding.EncodeToString(pub),
		Entries:       jsonEntries,
		Sessions:      jsonSessions,
		Budgets:       jsonBudgets,
	}

	digest, err := documentDigest(doc)
	if err != nil {
		return nil, err
	}

	doc.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(priv, digest))

	return doc, nil
}

// documentDigest returns the SHA-256 digest of the canonical JSON of the
// document with the signature field emptied, i.e. the bytes that get signed.
func documentDigest(doc *Document) ([]byte, error) {
	unsigned := *doc
	unsigned.Signature = ""

	canonical, err := canonicalJSON(unsigned)
	if err != nil {
		return nil, fmt.Errorf("canonical JSON: %w", err)
	}

	digest := sha256.Sum256(canonical)

	return digest[:], nil
}

func Verify(doc *Document) (VerifyResult, error) {
	res := VerifyResult{SchemaVersion: doc.SchemaVersion}

	pubBytes, err := base64.StdEncoding.DecodeString(doc.PublicKey)
	if err != nil {
		return res, fmt.Errorf("decode public key: %w", err)
	}

	if len(pubBytes) != ed25519.PublicKeySize {
		return res, fmt.Errorf("invalid public key size: got %d, want %d", len(pubBytes), ed25519.PublicKeySize)
	}

	pub := ed25519.PublicKey(pubBytes)

	sigBytes, err := base64.StdEncoding.DecodeString(doc.Signature)
	if err != nil {
		return res, fmt.Errorf("decode signature: %w", err)
	}

	var digest []byte

	switch doc.SchemaVersion {
	case SchemaVersion:
		digest, err = documentDigest(doc)
		if err != nil {
			return res, err
		}
	default:
		// The digest is taken over the canonical JSON of the parsed document,
		// so it depends on the fields the current types define. A document
		// written under an earlier schema loses that schema's fields and gains
		// the current ones as zero values, which changes the canonical form:
		// its signature cannot verify, and pretending otherwise would need a
		// parsing type per schema version.
		return res, fmt.Errorf(
			"unsupported schema version %q: only %s can be verified",
			doc.SchemaVersion, SchemaVersion,
		)
	}

	if !ed25519.Verify(pub, digest, sigBytes) {
		return res, errors.New("signature invalid: document has been tampered with")
	}

	return res, nil
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
