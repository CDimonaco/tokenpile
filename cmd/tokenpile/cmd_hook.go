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
	recorded := 0

	for _, turn := range turns {
		repo, issueNum := attribution.Resolve(bindings, turn.SessionID, turn.Cwd, turn.Branch)
		if repo == "" {
			repo = repoFromCwd(turn.Cwd)
		}

		if repo == "" {
			// Without a repository there is nothing to attribute to at all,
			// so the turn stays spooled rather than being recorded wrongly.
			continue
		}

		entry := usage.Entry{
			ID:        turn.ID,
			Repo:      repo,
			IssueNum:  issueNum,
			Agent:     turn.Agent,
			Model:     turn.Model,
			Usage:     turn.Usage,
			Source:    usage.SourceMeasured,
			SessionID: turn.SessionID,
			At:        turn.At,
		}

		if err = s.LogUsage(c.Context, entry); err != nil {
			slog.Warn("record spooled turn", "id", turn.ID, "err", err)

			return recorded
		}

		recorded++
	}

	if recorded == len(turns) {
		if err = spool.Clear(); err != nil {
			slog.Warn("clear spool", "err", err)
		}
	}

	return recorded
}

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
