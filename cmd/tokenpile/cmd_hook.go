package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

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
		captured = make(map[sessionKey]*usage.Session)
		cwds     = make(map[sessionKey]string)
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

		trackSession(captured, cwds, turn, repo, issueNum)
	}

	recordSessions(c, s, bindings, captured, cwds)

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

type sessionKey struct{ id, repo string }

// trackSession widens the span of the captured session a turn belongs to.
// Turns without a session id cannot be grouped and get no session row.
func trackSession(
	captured map[sessionKey]*usage.Session,
	cwds map[sessionKey]string,
	turn capture.Turn,
	repo string,
	issueNum *int,
) {
	if turn.SessionID == "" {
		return
	}

	key := sessionKey{id: turn.SessionID, repo: repo}

	sess, ok := captured[key]
	if !ok {
		captured[key] = &usage.Session{
			ID:             turn.SessionID,
			Repo:           repo,
			IssueNum:       issueNum,
			StartedAt:      turn.At,
			LastActivityAt: turn.At,
		}
		cwds[key] = turn.Cwd

		return
	}

	if turn.At.Before(sess.StartedAt) {
		sess.StartedAt = turn.At
	}

	if turn.At.After(sess.LastActivityAt) {
		sess.LastActivityAt = turn.At
	}

	if issueNum != nil {
		sess.IssueNum = issueNum
	}
}

// recordSessions writes one session row per captured session, not per turn:
// every Stop re-reads the whole transcript. Failures are logged and skipped,
// since a session row must never block recording usage.
func recordSessions(
	c *cli.Context,
	s store.Store,
	bindings *attribution.Store,
	captured map[sessionKey]*usage.Session,
	cwds map[sessionKey]string,
) {
	for key, sess := range captured {
		if err := s.UpsertCapturedSession(c.Context, *sess); err != nil {
			slog.Warn("record captured session", "id", sess.ID, "err", err)

			continue
		}

		binding, ok := bindings.Lookup(sess.ID, cwds[key])
		if !ok || (binding.Note == "" && len(binding.Tags) == 0) {
			continue
		}

		var note *string
		if binding.Note != "" {
			note = &binding.Note
		}

		if err := s.UpdateSessionAnnotations(c.Context, sess.ID, note, binding.Tags); err != nil {
			slog.Warn("copy binding annotations", "id", sess.ID, "err", err)
		}
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
