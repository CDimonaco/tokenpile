## Why

tokenpile is moving from token counts self-reported by the model to counts read from the agent's own transcript. That capture path needs, per agent, a deterministic trigger and a readable record of what the provider actually billed. Claude Code and opencode both provide one, verified against real local data. codex does not, on two counts: its own documentation states that "the transcript format isn't a stable interface for hooks and may change over time", and nothing about its transcript contents has been confirmed.

Keeping codex on the supported list would therefore mean shipping two tiers of support: agents whose numbers come from the provider, and one whose numbers come from a model guessing at its own context window. That guess is currently wrong by roughly two orders of magnitude. A supported agent that produces knowingly wrong numbers is worse than an unsupported one, because the error arrives under the same label, in the same reports, and inside the same signed export.

codex support is removed outright rather than deprecated. There are no users to migrate, and re-adding it later is a small change once its transcript is verified.

## What Changes

- **BREAKING** `codex` is no longer a supported agent. It is removed from the agent list, so `tokenpile skill install codex` fails with `ErrUnsupportedAgent` and `tokenpile skill list` no longer offers it.
- The embedded `internal/skill/templates/codex.md` template is deleted.
- No retirement or cleanup mechanism is introduced. Files a previous version installed at `~/.codex/skills/tokenpile/SKILL.md`, and any legacy tokenpile block in `~/.codex/AGENTS.md`, are left in place and are no longer reachable by `skill uninstall` or `reset`. This is a deliberate, accepted consequence.
- Docs and specs stop naming codex as supported.

## Capabilities

### Modified Capabilities

- `agent-skill`: the supported agent set becomes `claude-code` and `opencode`. The "Skill install location" requirement no longer enumerates codex; a new requirement states that an unsupported agent name is rejected.

## Impact

- `internal/skill/skill.go`: the codex entry is removed from `agents`, together with its `InstallPath` and `LegacySharedPath`; the `codexTemplate` embed directive is removed.
- `internal/skill/templates/codex.md`: deleted.
- `internal/skill/` tests: codex cases removed; a case asserting `ErrUnsupportedAgent` for `codex` is added so the removal is pinned by a test rather than only by absence.
- `cmd/tokenpile/cmd_reset.go`: no code change — it iterates `skill.List()`, which simply stops yielding codex.
- README: supported agents list.
- No storage, pricing or export change.
