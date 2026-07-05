package main

import (
	"crypto/ed25519"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/export"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

func exportCommands(s store.Store, priv ed25519.PrivateKey, version string) *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "export usage data as signed JSON",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "write to file instead of stdout",
			},
			&cli.StringFlag{
				Name:  "repo",
				Usage: "filter by repository (owner/repo)",
			},
			&cli.IntFlag{
				Name:  "issue",
				Usage: "filter by issue number",
			},
			&cli.StringFlag{
				Name:  "agent",
				Usage: "filter by agent name",
			},
			&cli.StringFlag{
				Name:  "model",
				Usage: "filter by model name",
			},
			&cli.StringFlag{
				Name:  "from",
				Usage: "start date filter (RFC3339)",
			},
			&cli.StringFlag{
				Name:  "to",
				Usage: "end date filter (RFC3339)",
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:  "verify",
				Usage: "verify a signed export file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "path to export file",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "pubkey",
						Usage: "expected signer public key: base64 string or path to a PEM/base64 file",
					},
				},
				Action: runVerify,
			},
		},
		Action: func(c *cli.Context) error {
			return runExport(c, s, priv, version)
		},
	}
}

func runVerify(c *cli.Context) error {
	data, err := os.ReadFile(c.String("file"))
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var doc export.Document
	if err = json.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("parse export: %w", err)
	}

	originChecked := false

	if pubkeyArg := c.String("pubkey"); pubkeyArg != "" {
		expected, keyErr := parseExpectedPubKey(pubkeyArg)
		if keyErr != nil {
			return fmt.Errorf("parse --pubkey: %w", keyErr)
		}

		docKey, decErr := base64.StdEncoding.DecodeString(doc.PublicKey)
		if decErr != nil {
			return fmt.Errorf("decode document public key: %w", decErr)
		}

		if len(docKey) != len(expected) || subtle.ConstantTimeCompare(expected, docKey) != 1 {
			fmt.Fprintln(c.App.Writer, "INVALID: public key mismatch")

			return errors.New("verification failed: document is not signed by the expected key")
		}

		originChecked = true
	}

	res, err := export.Verify(&doc)
	if err != nil {
		fmt.Fprintf(c.App.Writer, "INVALID: %v\n", err)

		return fmt.Errorf("verification failed: %w", err)
	}

	if res.Legacy {
		fmt.Fprintf(c.App.Writer, "OK: signature valid (schema %s, legacy), %d entries\n",
			res.SchemaVersion, len(doc.Entries))
		fmt.Fprintln(c.App.Writer,
			"WARNING: schema 2.0 signatures cover entries only; sessions and budgets are not protected")
	} else {
		fmt.Fprintf(c.App.Writer, "OK: signature valid (schema %s, full document), %d entries\n",
			res.SchemaVersion, len(doc.Entries))
	}

	if originChecked {
		fmt.Fprintln(c.App.Writer, "Origin verified: public key matches the expected key")
	} else {
		fmt.Fprintln(c.App.Writer,
			"Origin not verified: checked against the embedded key only (consistency, not provenance)")
	}

	return nil
}

// parseExpectedPubKey accepts a base64-encoded Ed25519 public key or a path
// to a file containing one, either as a PEM "ED25519 PUBLIC KEY" block
// (the identity.pub format) or as base64 text.
func parseExpectedPubKey(value string) (ed25519.PublicKey, error) {
	if raw, err := base64.StdEncoding.DecodeString(value); err == nil && len(raw) == ed25519.PublicKeySize {
		return ed25519.PublicKey(raw), nil
	}

	data, err := os.ReadFile(value)
	if err != nil {
		return nil, fmt.Errorf("read pubkey file: %w", err)
	}

	if block, _ := pem.Decode(data); block != nil {
		if len(block.Bytes) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("invalid public key size: got %d, want %d", len(block.Bytes), ed25519.PublicKeySize)
		}

		return ed25519.PublicKey(block.Bytes), nil
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, fmt.Errorf("decode pubkey file: %w", err)
	}

	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: got %d, want %d", len(raw), ed25519.PublicKeySize)
	}

	return ed25519.PublicKey(raw), nil
}

func runExport(c *cli.Context, s store.Store, priv ed25519.PrivateKey, version string) error {
	ctx := c.Context
	filter := usage.Filter{
		Repo:     strings.ToLower(c.String("repo")),
		IssueNum: c.Int("issue"),
		Agent:    c.String("agent"),
		Model:    c.String("model"),
	}

	if fromStr := c.String("from"); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return fmt.Errorf("parse --from: %w", err)
		}

		filter.From = &t
	}

	if toStr := c.String("to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return fmt.Errorf("parse --to: %w", err)
		}

		filter.To = &t
	}

	entries, err := s.ListEntries(ctx, filter)
	if err != nil {
		return fmt.Errorf("list entries: %w", err)
	}

	var sessions []usage.Session
	if filter.Repo != "" && filter.IssueNum > 0 {
		sessions, err = s.ListSessions(ctx, filter.Repo, filter.IssueNum)
		if err != nil {
			return fmt.Errorf("list sessions: %w", err)
		}
	}

	var budgets []export.IssueBudget
	if filter.Repo != "" && filter.IssueNum > 0 {
		b, budgetErr := s.GetBudget(ctx, filter.Repo, filter.IssueNum)
		if budgetErr != nil && !errors.Is(budgetErr, store.ErrBudgetNotFound) {
			return fmt.Errorf("get budget: %w", budgetErr)
		}

		if b != nil {
			budgets = []export.IssueBudget{{Repo: filter.Repo, IssueNum: filter.IssueNum, Amount: *b}}
		}
	}

	doc, err := export.Build(entries, sessions, budgets, priv, version)
	if err != nil {
		return fmt.Errorf("build export: %w", err)
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal export: %w", err)
	}

	if outPath := c.String("output"); outPath != "" {
		if err = os.WriteFile(outPath, data, 0o600); err != nil {
			return fmt.Errorf("write output: %w", err)
		}

		fmt.Fprintf(c.App.Writer, "Exported %d entries to %s\n", len(entries), outPath)

		return nil
	}

	fmt.Fprintf(c.App.Writer, "%s\n", data)

	return nil
}
