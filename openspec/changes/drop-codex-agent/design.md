## Context

`internal/skill/skill.go` holds a package-level `agents` slice of three `Agent` values, each carrying an embedded template and path functions. `List()` returns that slice; `Install`, `Uninstall` and `IsInstalled` look an agent up by name and return `ErrUnsupportedAgent` when it is absent. `cmd_reset.go` iterates `skill.List()` twice: once to enumerate what will be deleted, once to delete it.

The codex entry additionally carries a `LegacySharedPath` pointing at `~/.codex/AGENTS.md`, used by `cleanupLegacy` to strip a marked tokenpile block left by a pre-`SKILL.md` version.

## Goals / Non-Goals

**Goals:**
- Stop claiming support for an agent whose token counts cannot be measured.
- Keep the removal small enough that re-adding codex later is a matter of restoring one struct and one template.
- Pin the removal with a test, so a future refactor cannot silently reintroduce a half-supported agent.

**Non-Goals:**
- Any retirement, deprecation-warning or cleanup mechanism for already-installed codex files.
- Verifying what codex's transcript actually contains. That verification belongs to whatever change re-adds support.
- Touching the `claude-code` or `opencode` entries.

## Decisions

**1. Outright removal, not a retired-agent list.**
The natural-looking alternative was a `retiredAgents` slice excluded from `Install`/`List` but still honoured by `Uninstall` and `reset`, mirroring the existing `LegacyDedicatedPath`/`LegacySharedPath` cleanup pattern one level up. It was considered and rejected on the owner's explicit instruction: there are no users, so the mechanism would exist solely to tidy files on machines that do not exist. The cost of the decision is stated plainly in the proposal rather than engineered around.

**2. Orphaned files are accepted, not hidden.**
After this change, `~/.codex/skills/tokenpile/SKILL.md` and any legacy block in `~/.codex/AGENTS.md` become unreachable by tokenpile: `reset` will no longer remove them, because `reset` only knows the agents `List()` yields. This is recorded here so the next person to read `reset`'s deletion scope does not conclude it is a bug.

**3. The removal is pinned by an assertion, not by absence.**
`Install("codex")` returning `ErrUnsupportedAgent` gets its own test. Without it, the only thing preventing codex from drifting back is that nobody adds it, and the reason it was dropped lives in a proposal rather than in the suite.

**4. codex keeps no special status in the error path.**
It is rejected exactly like any unknown string. A dedicated "codex is no longer supported" message was considered and rejected: it implies a migration story for users who do not exist, and it would have to be maintained until someone decides it is stale.

## Risks / Trade-offs

- [A returning codex user finds their installed skill still present and silently stale] → accepted; the skill only instructs an agent to call a CLI, so the failure mode is a command that reports an unsupported agent, not corrupt data.
- [`reset` no longer leaves the machine completely clean] → accepted and documented here; the reset spec's claim is about tokenpile's own state for supported agents.
- [Dropping an agent could look like a capability regression] → it removes a capability that was measurably wrong; the proposal states the reasoning so the record survives the commit.
