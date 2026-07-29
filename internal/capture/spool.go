package capture

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Spool is an append-only journal of captured turns awaiting storage.
//
// It exists because a hook that fails loses data silently. In Claude Code's
// hook semantics, exit 2 blocks the action and any other non-zero exit merely
// prints to stderr and carries on, so a capture path that wrote straight to
// SQLite would drop turns whenever the database was busy, mid-upgrade or on a
// full disk — at exactly the moment nobody is watching. Appending one line to a
// file is close to unfailable and needs no database.
type Spool struct {
	path string
}

func NewSpool(path string) *Spool {
	return &Spool{path: path}
}

// Path is where the spool lives.
func (s *Spool) Path() string { return s.path }

// Append writes turns to the journal. Each line is self-contained, so a
// truncated final write costs one turn rather than the file.
func (s *Spool) Append(turns []Turn) error {
	if len(turns) == 0 {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create spool directory: %w", err)
	}

	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open spool: %w", err)
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	for _, t := range turns {
		if err = enc.Encode(t); err != nil {
			return fmt.Errorf("write spool record: %w", err)
		}
	}

	return f.Sync()
}

// AppendRaw preserves a payload that could not be parsed, so a format change
// costs an investigation rather than the measurements themselves.
func (s *Spool) AppendRaw(agent string, payload []byte, reason error) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create spool directory: %w", err)
	}

	path := s.path + ".raw"

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open raw spool: %w", err)
	}
	defer func() { _ = f.Close() }()

	record := map[string]any{
		"at":      time.Now().UTC().Format(time.RFC3339),
		"agent":   agent,
		"reason":  reason.Error(),
		"payload": string(payload),
	}

	return json.NewEncoder(f).Encode(record)
}

// Read returns everything currently spooled. Malformed lines are skipped rather
// than aborting the drain: one bad record must not strand every good one behind
// it.
func (s *Spool) Read() ([]Turn, error) {
	f, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("open spool: %w", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var turns []Turn

	for scanner.Scan() {
		var t Turn
		if err = json.Unmarshal(scanner.Bytes(), &t); err != nil {
			continue
		}

		turns = append(turns, t)
	}

	if err = scanner.Err(); err != nil {
		return turns, fmt.Errorf("read spool: %w", err)
	}

	return turns, nil
}

// Clear empties the journal after a successful drain. It is only safe because
// storage is idempotent on the turn id: a crash between storing and clearing
// replays records that the store then ignores.
func (s *Spool) Clear() error {
	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("clear spool: %w", err)
	}

	return nil
}

// Pending reports how many records are waiting, for surfacing a backlog.
func (s *Spool) Pending() int {
	turns, err := s.Read()
	if err != nil {
		return 0
	}

	return len(turns)
}
