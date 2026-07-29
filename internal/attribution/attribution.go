// Package attribution decides which GitHub issue a captured turn belongs to.
//
// Attribution is deliberately separate from capture. Capture must never fail
// for want of an issue number: an unattributed measurement can be assigned
// later, a discarded one is gone. Everything here is therefore best-effort, and
// producing no answer is a valid outcome.
package attribution

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

// Binding records that a session or working directory belongs to an issue.
type Binding struct {
	Repo     string    `json:"repo"`
	IssueNum int       `json:"issue_num"`
	Note     string    `json:"note,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	BoundAt  time.Time `json:"bound_at"`
}

// Store persists bindings keyed by session id and by working directory. A
// session id is the precise key; the directory is the fallback for the common
// case where the user binds before the agent's session is known.
type Store struct {
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

type bindingFile struct {
	Sessions    map[string]Binding `json:"sessions,omitempty"`
	Directories map[string]Binding `json:"directories,omitempty"`
}

func (s *Store) load() (bindingFile, error) {
	var f bindingFile

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return bindingFile{Sessions: map[string]Binding{}, Directories: map[string]Binding{}}, nil
	}

	if err != nil {
		return f, fmt.Errorf("read bindings: %w", err)
	}

	if err = json.Unmarshal(data, &f); err != nil {
		return f, fmt.Errorf("parse bindings: %w", err)
	}

	if f.Sessions == nil {
		f.Sessions = map[string]Binding{}
	}

	if f.Directories == nil {
		f.Directories = map[string]Binding{}
	}

	return f, nil
}

func (s *Store) save(f bindingFile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create binding directory: %w", err)
	}

	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal bindings: %w", err)
	}

	if err = os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write bindings: %w", err)
	}

	return nil
}

// Bind records a binding. Binding again replaces the previous value: the user
// has moved on to another issue, and only turns captured afterwards should
// follow.
func (s *Store) Bind(sessionID, directory string, b Binding) error {
	f, err := s.load()
	if err != nil {
		return err
	}

	b.BoundAt = time.Now().UTC()

	if sessionID != "" {
		f.Sessions[sessionID] = b
	}

	if directory != "" {
		f.Directories[directory] = b
	}

	return s.save(f)
}

// Lookup returns the binding for a session, falling back to its directory.
func (s *Store) Lookup(sessionID, directory string) (Binding, bool) {
	f, err := s.load()
	if err != nil {
		return Binding{}, false
	}

	if b, ok := f.Sessions[sessionID]; ok && sessionID != "" {
		return b, true
	}

	if b, ok := f.Directories[directory]; ok && directory != "" {
		return b, true
	}

	return Binding{}, false
}

// Clear removes a session's binding.
func (s *Store) Clear(sessionID string) error {
	f, err := s.load()
	if err != nil {
		return err
	}

	delete(f.Sessions, sessionID)

	return s.save(f)
}

// branchIssue matches the issue number conventionally embedded in a branch
// name: a leading number, or one after a type prefix or an "issue" marker.
//
// Deliberately offline. Resolving a branch to a pull request to a linked issue
// through the GitHub API would make every captured turn depend on network and
// credentials, which contradicts a capture path that must not fail.
var branchIssue = regexp.MustCompile(`(?:^|/|-|_)(?:issue[-_]?)?(\d{1,6})(?:$|[-_/])`)

// InferFromBranch extracts an issue number from a branch name, or reports that
// it cannot. A branch like "main" or "refactor/pricing" yields nothing, which
// is a normal outcome rather than a failure.
func InferFromBranch(branch string) (int, bool) {
	if branch == "" {
		return 0, false
	}

	m := branchIssue.FindStringSubmatch(branch)
	if m == nil {
		return 0, false
	}

	n, err := strconv.Atoi(m[1])
	if err != nil || n <= 0 {
		return 0, false
	}

	return n, true
}

// Resolve applies the resolution order: an explicit binding, then the branch
// name, then nothing. The third outcome is the point of the design — it is why
// capture can proceed without an issue.
func Resolve(s *Store, sessionID, directory, branch string) (string, *int) {
	if b, ok := s.Lookup(sessionID, directory); ok {
		// A binding always carries the repository, but an issue of zero is not
		// an attribution: it would record every turn against issue #0.
		if b.IssueNum > 0 {
			n := b.IssueNum

			return b.Repo, &n
		}

		if n, inferred := InferFromBranch(branch); inferred {
			return b.Repo, &n
		}

		return b.Repo, nil
	}

	if n, ok := InferFromBranch(branch); ok {
		return "", &n
	}

	return "", nil
}
