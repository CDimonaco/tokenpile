package attribution_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cdimonaco/tokenpile/internal/attribution"
)

func newStore(t *testing.T) *attribution.Store {
	t.Helper()

	return attribution.NewStore(filepath.Join(t.TempDir(), "bindings.json"))
}

func TestInferFromBranch(t *testing.T) {
	tests := []struct {
		branch string
		want   int
		ok     bool
	}{
		{branch: "feat/42-borrow-gh-credential", want: 42, ok: true},
		{branch: "42-quick-fix", want: 42, ok: true},
		{branch: "fix/issue-7", want: 7, ok: true},
		{branch: "bugfix/issue_128_crash", want: 128, ok: true},
		{branch: "chore/1234", want: 1234, ok: true},
		{branch: "main", ok: false},
		{branch: "refactor/pricing", ok: false},
		{branch: "", ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.branch, func(t *testing.T) {
			got, ok := attribution.InferFromBranch(tc.branch)

			assert.Equal(t, tc.ok, ok)

			if tc.ok {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestBindAndLookupBySession(t *testing.T) {
	s := newStore(t)

	require.NoError(t, s.Bind("sess-1", "", attribution.Binding{Repo: "o/r", IssueNum: 42}))

	b, ok := s.Lookup("sess-1", "")

	require.True(t, ok)
	assert.Equal(t, 42, b.IssueNum)
	assert.Equal(t, "o/r", b.Repo)
}

func TestBindAndLookupByDirectory(t *testing.T) {
	s := newStore(t)

	require.NoError(t, s.Bind("", "/w/repo", attribution.Binding{Repo: "o/r", IssueNum: 9}))

	b, ok := s.Lookup("unknown-session", "/w/repo")

	require.True(t, ok)
	assert.Equal(t, 9, b.IssueNum)
}

func TestRebindReplaces(t *testing.T) {
	s := newStore(t)

	require.NoError(t, s.Bind("sess-1", "", attribution.Binding{Repo: "o/r", IssueNum: 42}))
	require.NoError(t, s.Bind("sess-1", "", attribution.Binding{Repo: "o/r", IssueNum: 7}))

	b, ok := s.Lookup("sess-1", "")

	require.True(t, ok)
	assert.Equal(t, 7, b.IssueNum)
}

// An explicit binding is a statement of intent; a branch name is a guess.
func TestResolve_BindingWinsOverBranch(t *testing.T) {
	s := newStore(t)
	require.NoError(t, s.Bind("sess-1", "", attribution.Binding{Repo: "o/r", IssueNum: 42}))

	repo, issue := attribution.Resolve(s, "sess-1", "/w/repo", "feat/7-other")

	require.NotNil(t, issue)
	assert.Equal(t, 42, *issue)
	assert.Equal(t, "o/r", repo)
}

func TestResolve_FallsBackToBranch(t *testing.T) {
	s := newStore(t)

	_, issue := attribution.Resolve(s, "sess-1", "/w/repo", "feat/42-thing")

	require.NotNil(t, issue)
	assert.Equal(t, 42, *issue)
}

// The outcome that makes the whole design work: no issue is not an error.
func TestResolve_NoIssueIsValid(t *testing.T) {
	s := newStore(t)

	repo, issue := attribution.Resolve(s, "sess-1", "/w/repo", "main")

	assert.Nil(t, issue)
	assert.Empty(t, repo)
}

func TestClearRemovesBinding(t *testing.T) {
	s := newStore(t)
	require.NoError(t, s.Bind("sess-1", "", attribution.Binding{Repo: "o/r", IssueNum: 42}))

	require.NoError(t, s.Clear("sess-1"))

	_, ok := s.Lookup("sess-1", "")
	assert.False(t, ok)
}

func TestLookupMissingFileIsNotAnError(t *testing.T) {
	s := attribution.NewStore(filepath.Join(t.TempDir(), "does-not-exist.json"))

	_, ok := s.Lookup("sess-1", "/w/repo")

	assert.False(t, ok)
}

// A binding names the repository even when it names no usable issue. That
// combination is what lets a turn be recorded against the right repo while
// staying unattributed.
func TestResolve_BindingWithoutIssueStillGivesRepo(t *testing.T) {
	s := newStore(t)
	require.NoError(t, s.Bind("sess-1", "", attribution.Binding{Repo: "o/r", IssueNum: 0}))

	repo, issue := attribution.Resolve(s, "sess-1", "", "main")

	assert.Equal(t, "o/r", repo)
	assert.Nil(t, issue, "issue zero is not an attribution")
}
