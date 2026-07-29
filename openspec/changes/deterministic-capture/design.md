## Context

Verified locally rather than assumed:

**Claude Code.** Hooks live in `~/.claude/settings.json` (also project and local settings) and persist for the whole session. `Stop` fires when Claude finishes responding. Every hook receives JSON on stdin with `session_id`, `prompt_id`, `transcript_path`, `cwd`, `permission_mode` and `hook_event_name`. No hook payload carries token counts, and `model` appears only on `SessionStart`. The transcript at `transcript_path` is JSONL whose assistant messages carry `message.usage` with `input_tokens`, `cache_creation_input_tokens`, `cache_read_input_tokens` and `output_tokens`, plus `message.model`. There is no separate reasoning field: Anthropic bills thinking tokens inside output.

**opencode.** Plugins are TypeScript, loaded from `~/.config/opencode/plugin/` or a project `.opencode/plugins/`, and receive a context with `project`, `directory`, `worktree`, `client` and `$`. The `session.idle` event fires when a session goes idle. The `message` table stores one JSON blob per message containing `tokens: {total, input, output, reasoning, cache: {write, read}}`, `cost`, `modelID`, `providerID` and `path.cwd`.

**Rejected for a reason worth recording.** Claude Code's `SKILL.md` frontmatter accepts a `hooks` field, which looked ideal since `skill install` already writes that file. The documentation is explicit that such hooks "are scoped to the component's lifecycle and only run when that component is active" — that is, only while the model has chosen to load the skill. Using them would reintroduce exactly the non-determinism this change removes. Hooks must go in `settings.json`.

## Goals / Non-Goals

**Goals:**
- Token counts that come from the provider, never from a model's estimate.
- A trigger that fires whether or not the model cooperates.
- No silent data loss when the capture path fails.
- Attribution that can be wrong or absent without costing a measurement.

**Non-Goals:**
- codex support; it is removed by `drop-codex-agent` and its transcript format is documented as unstable.
- Reconstructing history from transcripts already on disk. The readers make it possible later; this change captures forward only.
- Reconciling tokenpile's computed cost against opencode's own `cost` field. Cost stays computed from tokenpile's pricing so all agents are comparable.
- Making cost meaningful under subscription billing. It remains a list-price figure.

## Decisions

**1. `issue_num` becomes nullable. This is the change.**
Everything else follows. With the constraint in place, capture must know the issue, which forces the model into the loop, which is where the estimates came from. Nullable turns attribution from a precondition into an annotation.
The cost is that every query touching `issue_num` must handle null, and reports must decide how to present usage that belongs to no issue. That is real work and it is the price of the whole design.

**2. Capture writes to an append-only spool first, then reconciles.**
In Claude Code's hook semantics, exit 2 blocks the action and any other non-zero exit prints stderr and continues. A hook that fails — locked database, upgrade in progress, full disk — therefore loses a turn silently. Writing one JSON line to a spool file is close to unfailable and needs no database; a reconciler folds the spool into the store on the next tokenpile invocation.
Alternative considered: write directly to SQLite from the hook. Rejected — it makes every turn depend on a writable database at exactly the moment the user is not watching, and "deterministic trigger" would not mean "reliable record".

**3. Attribution resolves in a fixed order, and the last option is none.**
```
1. explicit binding for this session      tokenpile bind --issue 42
2. inference from the git branch          feat/42-foo → 42
3. none                                   entry recorded, issue null
```
Step 3 is the one that matters. Every prior design treated an unknown issue as a reason not to record; here it is a reason to record without one.
Branch inference stays offline and pattern-based. Resolving a branch to a pull request to a linked issue through the GitHub API was considered and rejected for this change: it makes capture depend on network and credentials on every turn, which contradicts the goal of a capture path that cannot fail.

**4. The skill declares, it no longer measures.**
The template loses its token estimation instructions entirely and gains one instruction: call `tokenpile bind --issue N` when the issue becomes known. This is the only part of the pipeline that must come from the model, and it is the part a model is actually competent at.
The residual non-determinism is now benign: a forgotten bind produces an unattributed measurement, recoverable at any time, rather than no measurement at all.

**5. Reconciliation is a separate surface, not an interactive prompt.**
Hooks cannot ask the user anything. Unattributed sessions are therefore listed after the fact — grouped by repo, branch and time window, with branch-derived suggestions — and assigned in bulk. Grouping matters: a session is dozens of turns, and assigning them one by one would make the feature unusable.

**6. Two readers, one internal shape.**
`internal/capture` defines the tier shape once and gives each agent a reader that maps its own format onto it. Claude Code's per-turn JSONL and opencode's per-message rows are both per-turn, so the earlier worry about mismatched granularity does not apply. opencode's `reasoning` maps directly; Claude Code always reports zero, which is correct rather than missing, since Anthropic bills thinking inside output.

**7. Hook installation rides with skill installation.**
`skill install <agent>` writes both the skill and the hook — a `Stop` entry merged into `~/.claude/settings.json`, or the plugin file for opencode. `skill uninstall` and `reset` remove both.
Merging into an existing `settings.json` must preserve every other hook and setting; this is the same class of surgery as the existing marked-block removal in shared files, and deserves the same care and the same tests.

## Risks / Trade-offs

- [Editing a user's `~/.claude/settings.json` can damage unrelated configuration] → merge, never rewrite; keep a marker identifying tokenpile's entry; test against a settings file containing foreign hooks.
- [The spool grows if reconciliation never runs] → reconcile on every tokenpile invocation and report the backlog in `report` output.
- [Transcript formats are not contracts] → both readers are tested against captured fixtures, and an unparseable transcript spools the raw payload rather than discarding the turn.
- [Unattributed usage complicates budgets] → a budget is per issue, so unattributed usage counts toward none; it must be visible somewhere or it becomes invisible spending. This needs an explicit answer before implementation.
- [Attribution can be wrong, not just missing] → branch inference is a guess; bulk assignment must be reversible.
- [Double counting if both the hook and the skill log] → the skill loses its logging instruction in this change; `log` remains for manual entries and is always `estimated`, so the two paths stay distinguishable by `source`.
- [The change is large] → the capture path stands alone recording everything unattributed. If reconciliation grows, it splits into a fourth change and this one still ships something correct.
