package main

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/config"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

func runUnattributedCmd(t *testing.T, s *store.SQLiteStore, args ...string) (string, error) {
	t.Helper()

	dir := t.TempDir()
	paths := config.Paths{
		SpoolPath:    filepath.Join(dir, "spool.jsonl"),
		BindingsPath: filepath.Join(dir, "bindings.json"),
	}

	var out bytes.Buffer

	app := &cli.App{
		Writer:   &out,
		Commands: []*cli.Command{unattributedCommand(s, paths)},
	}

	err := app.RunContext(context.Background(), append([]string{"tok"}, args...))

	return out.String(), err
}

func TestUnattributed_ListShowsBranchAndSuggestion(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "owner/repo", Agent: "a", Model: "m", SessionID: "sess-suggests", Branch: "feat/42-x",
		Usage: usage.Usage{InputFresh: 10}, Source: usage.SourceMeasured,
	}))
	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "owner/repo", Agent: "a", Model: "m", SessionID: "sess-main", Branch: "main",
		Usage: usage.Usage{InputFresh: 10}, Source: usage.SourceMeasured,
	}))

	out, err := runUnattributedCmd(t, s, "unattributed")
	require.NoError(t, err)

	assert.Contains(t, out, "feat/42-x")
	assert.Contains(t, out, "#42")
	assert.Contains(t, out, "main")
}

func TestUnattributed_AssignAcceptsTheSuggestion(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "owner/repo", Agent: "a", Model: "m", SessionID: "sess-1", Branch: "feat/42-x",
		Usage: usage.Usage{InputFresh: 10}, Source: usage.SourceMeasured,
	}))

	out, err := runUnattributedCmd(t, s, "unattributed", "assign", "sess-1")
	require.NoError(t, err)
	assert.Contains(t, out, "#42")

	entries, err := s.ListEntries(ctx, usage.Filter{Repo: "owner/repo", IssueNum: 42})
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

// Without a suggestion, omitting --issue must fail rather than pick something.
func TestUnattributed_AssignWithoutSuggestionRequiresIssue(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.LogUsage(ctx, usage.Entry{
		Repo: "owner/repo", Agent: "a", Model: "m", SessionID: "sess-1", Branch: "main",
		Usage: usage.Usage{InputFresh: 10}, Source: usage.SourceMeasured,
	}))

	_, err := runUnattributedCmd(t, s, "unattributed", "assign", "sess-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no issue suggested")

	groups, err := s.ListUnattributed(ctx, "owner/repo")
	require.NoError(t, err)
	assert.Len(t, groups, 1, "nothing was assigned")
}
