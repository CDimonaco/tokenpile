package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/cdimonaco/tokenpile/internal/attribution"
	"github.com/cdimonaco/tokenpile/internal/capture"
	"github.com/cdimonaco/tokenpile/internal/config"
	"github.com/cdimonaco/tokenpile/internal/provider"
	"github.com/cdimonaco/tokenpile/internal/store"
	"github.com/cdimonaco/tokenpile/internal/usage"
)

// hookCommand is invoked by an agent, not by a person. It reads that agent's
// hook payload from stdin, extracts the usage the provider actually reported,
// and spools it.
//
// It never fails loudly on the agent's behalf: a hook that returns an error
// mid-session is noise the user cannot act on. Failures are reported on stderr
// and the payload is preserved.
func hookCommand(paths config.Paths) *cli.Command {
	return &cli.Command{
		Name:      "hook",
		Usage:     "record usage from an agent hook (reads the hook payload on stdin)",
		ArgsUsage: "<agent>",
		Hidden:    true,
		Action: func(c *cli.Context) error {
			agent := c.Args().First()
			if agent == "" {
				return errors.New("agent name is required: tokenpile hook <agent>")
			}

			payload, err := io.ReadAll(c.App.Reader)
			if err != nil {
				return fmt.Errorf("read hook payload: %w", err)
			}

			spool := capture.NewSpool(paths.SpoolPath)

			turns, err := turnsFromPayload(agent, payload)
			if err != nil {
				// Preserve rather than discard: an agent changing its format
				// must cost an investigation, not the measurements.
				if rawErr := spool.AppendRaw(agent, payload, err); rawErr != nil {
					slog.Error("preserve unparseable payload", "err", rawErr)
				}

				if errors.Is(err, capture.ErrNoUsage) {
					return nil
				}

				return fmt.Errorf("read %s payload: %w", agent, err)
			}

			if err = spool.Append(turns); err != nil {
				return fmt.Errorf("spool turns: %w", err)
			}

			return nil
		},
	}
}

func turnsFromPayload(agent string, payload []byte) ([]capture.Turn, error) {
	switch agent {
	case capture.AgentClaudeCode:
		var hook capture.HookPayload
		if err := json.Unmarshal(payload, &hook); err != nil {
			return nil, fmt.Errorf("parse hook payload: %w", err)
		}

		if hook.TranscriptPath == "" {
			return nil, errors.New("hook payload carries no transcript_path")
		}

		return capture.ReadClaudeCodeTranscriptFile(hook.TranscriptPath)
	case capture.AgentOpenCode:
		return capture.ReadOpenCodePayload(newBytesReader(payload))
	default:
		return nil, fmt.Errorf("unsupported agent %q", agent)
	}
}

// reconcileSpool folds spooled turns into the store. It runs on ordinary
// tokenpile invocations rather than from the hook, so a busy or locked database
// delays recording instead of losing it.
//
// Storage is idempotent on the turn id, so replaying a spool that was already
// drained is harmless — which is what makes clearing it safe without a
// two-phase commit.
func reconcileSpool(c *cli.Context, s store.Store, paths config.Paths) int {
	spool := capture.NewSpool(paths.SpoolPath)

	turns, err := spool.Read()
	if err != nil || len(turns) == 0 {
		return 0
	}

	bindings := attribution.NewStore(paths.BindingsPath)

	var (
		recorded int
		handled  int
		unusable []capture.Turn
		spans    = map[string]*sessionSpan{}
		order    []string
	)

	for _, turn := range turns {
		repo, issueNum := attribution.Resolve(bindings, turn.SessionID, turn.Cwd, turn.Branch)
		if repo == "" {
			repo = repoFromCwd(turn.Cwd)
		}

		if repo == "" {
			// No repository can ever be derived for this turn: its working
			// directory is not a git checkout with an origin remote, and that
			// will not change. Quarantining it keeps the spool draining —
			// leaving it in place would stall every later turn behind it and
			// grow the file without bound — while still preserving the record.
			unusable = append(unusable, turn)
			handled++

			continue
		}

		entry := usage.Entry{
			ID:        turn.ID,
			Repo:      repo,
			IssueNum:  issueNum,
			Agent:     turn.Agent,
			Model:     turn.Model,
			Usage:     turn.Usage,
			Branch:    turn.Branch,
			Source:    usage.SourceMeasured,
			SessionID: turn.SessionID,
			At:        turn.At,
		}

		if err = s.LogUsage(c.Context, entry); err != nil {
			slog.Warn("record spooled turn", "id", turn.ID, "err", err)

			return recorded
		}

		recorded++
		handled++

		trackSession(spans, &order, turn, repo, issueNum)
	}

	// Entries are the source of truth: a failed session write only warns and
	// never keeps the spool from being cleared.
	for _, id := range order {
		recordSession(c.Context, s, bindings, id, spans[id])
	}

	for _, turn := range unusable {
		payload, marshalErr := json.Marshal(turn)
		if marshalErr != nil {
			continue
		}

		if rawErr := spool.AppendRaw(turn.Agent, payload, errNoRepo); rawErr != nil {
			slog.Warn("quarantine unattributable turn", "id", turn.ID, "err", rawErr)
		}
	}

	if handled == len(turns) {
		if err = spool.Clear(); err != nil {
			slog.Warn("clear spool", "err", err)
		}
	}

	return recorded
}

// sessionSpan accumulates what reconcileSpool learns about one agent session.
type sessionSpan struct {
	repo     string
	cwd      string
	issueNum *int
	first    time.Time
	last     time.Time
}

func trackSession(spans map[string]*sessionSpan, order *[]string, turn capture.Turn, repo string, issueNum *int) {
	if turn.SessionID == "" {
		return
	}

	span, ok := spans[turn.SessionID]
	if !ok {
		spans[turn.SessionID] = &sessionSpan{
			repo: repo, cwd: turn.Cwd, issueNum: issueNum, first: turn.At, last: turn.At,
		}
		*order = append(*order, turn.SessionID)

		return
	}

	if turn.At.Before(span.first) {
		span.first = turn.At
	}

	if turn.At.After(span.last) {
		span.last = turn.At
	}

	if issueNum != nil {
		span.issueNum = issueNum
	}
}

// recordSession upserts the session row and copies the binding's note and tags
// onto it.
func recordSession(ctx context.Context, s store.Store, bindings *attribution.Store, id string, span *sessionSpan) {
	if err := s.UpsertCapturedSession(ctx, id, span.repo, span.issueNum, span.first, span.last); err != nil {
		slog.Warn("record session", "session", id, "err", err)

		return
	}

	if binding, ok := bindings.Lookup(id, span.cwd); ok {
		applyAnnotations(ctx, s, id, binding.Note, binding.Tags)
	}
}

// errNoRepo explains why a turn was quarantined rather than recorded.
var errNoRepo = errors.New("no repository could be derived from the turn's working directory")

func newBytesReader(b []byte) io.Reader { return bytes.NewReader(b) }

// repoFromCwd infers the repository from the directory the turn happened in,
// which the agent records in its own transcript.
func repoFromCwd(cwd string) string {
	if cwd == "" {
		return ""
	}

	repo, err := provider.InferRepoIn(cwd)
	if err != nil {
		return ""
	}

	return repo
}
