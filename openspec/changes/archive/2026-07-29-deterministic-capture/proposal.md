## Why

tokenpile's token counts come from a skill that asks the agent to estimate its own usage at roughly four characters per token. On a real session that estimate came out around 61x below the truth, because a model cannot see its own context: system prompt, tool definitions, and above all the prompt cache, which accounted for 98.4% of the tokens actually consumed.

The estimate is not the root cause. `usage_entries.issue_num` is `NOT NULL`, so nothing can be recorded without first knowing the issue. Only the model knows the issue, so the model had to do the logging, so the model had to supply the numbers too. A two-order-of-magnitude error descends from a schema constraint:

```
issue_num NOT NULL
  → cannot record a token without knowing the issue
     → only the model knows the issue
        → the model must do the logging
           → the model must supply the numbers
              → the numbers are guesses
```

Both agents that remain supported already write the true numbers to disk, as reported by the provider, and both offer a deterministic trigger. Verified locally: Claude Code fires a `Stop` hook per turn and passes `transcript_path`, whose JSONL carries `input_tokens`, `cache_creation_input_tokens`, `cache_read_input_tokens` and `output_tokens` per assistant message; opencode fires `session.idle` to a plugin, and its `message` table stores `{"tokens":{"input","output","reasoning","cache":{"write","read"}}}` per message alongside `cost`, `modelID` and `cwd`.

Making `issue_num` nullable inverts the relationship between the two problems. Capture becomes unconditional and lossless; attribution becomes a separate, deferrable, correctable act. Today a missing attribution means a lost measurement; afterwards it means an unattributed one.

## What Changes

- **BREAKING** `issue_num` becomes nullable on usage entries and sessions. This is the pivot the rest of the change depends on.
- New `tokenpile hook claude-code`: reads the hook JSON from stdin, parses the transcript at `transcript_path`, and records the turn's per-tier usage with `source = measured`. Installed as a `Stop` hook in `~/.claude/settings.json`.
- New opencode plugin, installed alongside the skill, subscribing to `session.idle` and invoking tokenpile with the per-message usage read through the opencode client.
- Hooks SHALL write to an append-only spool first and reconcile into the store afterwards. A hook that fails mid-write must lose nothing: in Claude Code's hook semantics a non-zero exit other than 2 prints stderr and carries on, so an unspooled failure is silent data loss.
- Attribution is resolved in order at capture time: an explicit binding for the session, then inference from the git branch name, then nothing — in which case the entry is recorded unattributed rather than dropped.
- New `tokenpile bind --issue N`, associating the current session or working directory with an issue.
- New reconciliation surface listing unattributed sessions grouped by repo, branch and time window, with branch-derived suggestions, assignable in bulk.
- **BREAKING** The agent skill stops logging. Its only remaining job is declaring the issue via `bind` — the one thing a model reliably knows. The skill's residual non-determinism becomes harmless: forgetting to bind costs attribution, not measurement.
- Hook installation is part of `tokenpile skill install`, and removal part of `skill uninstall` and `reset`.

## Capabilities

### New Capabilities

- `capture`: the hook and plugin capture path — triggers, transcript reading, the spool, provenance, and installation of the hook alongside the skill.
- `attribution`: binding an issue to a session, inferring it from the branch, recording entries unattributed, and reconciling them later.

### Modified Capabilities

- `usage-tracking`: entries may have no issue; `log` remains available for manual and estimated entries.
- `sessions`: a session's issue becomes optional and assignable after the fact.
- `agent-skill`: the skill declares the issue instead of logging usage; installation also installs the capture hook.

## Impact

- `internal/store/`: `issue_num` nullable on `usage_entries` and `sessions`; queries and aggregations handle the null case; new methods for unattributed listings and bulk assignment.
- `internal/usage/`: `Entry.IssueNum` and `Session.IssueNum` become `*int`.
- New `internal/capture/`: hook payload types, Claude Code transcript reader, opencode reader, spool writer and reconciler.
- New `internal/attribution/`: binding store, branch-name inference.
- `internal/skill/`: hook installation and removal per agent; opencode plugin file as a second embedded artifact.
- `cmd/tokenpile/`: `cmd_hook.go`, `cmd_bind.go`, reconciliation entry point; `cmd_reset.go` removes hooks too.
- `internal/tui/`: unattributed sessions view.
- Reports and budgets must decide how unattributed usage appears; it exists but belongs to no issue.
- Depends on `token-tier-accounting`: without per-tier columns there is nowhere to write what the transcripts contain.
