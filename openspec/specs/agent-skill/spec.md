## Purpose

Define the skill files tokenpile installs into coding agents (Claude Code, opencode) so an agent declares which issue it is working on, and how those files and their capture hooks are placed, refreshed and removed. Token counts come from the capture path, not from the skill.

## Requirements

### Requirement: Skill file content

The skill template SHALL NOT instruct the agent to estimate or report token counts. Token counts come from the capture path, which reads what the provider reported; a model cannot observe its own context or cache and any figure it supplies is a guess.

The skill's remaining responsibilities are to declare the issue being worked on by calling `tokenpile bind --issue N` once it is known, and to answer questions about usage data by running CLI commands. The template SHALL instruct the agent to pass `--note` and `--tag` when binding. The note SHALL be a single line (max ~100 chars) summarizing what is being worked on. Tags SHALL be chosen from a documented vocabulary: `refactor`, `debug`, `feature`, `test`, `docs`, `spike`, `review`.

Failing to bind SHALL cost attribution only: usage is still captured, unattributed, and can be assigned later.

#### Scenario: Skill file provides correct CLI invocation
- **WHEN** the skill template for any agent is rendered
- **THEN** the example invocation is `tokenpile bind` with `--issue`, `--note` and `--tag` flags
- **THEN** the template documents the recommended tag vocabulary

#### Scenario: Template does not ask for token counts
- **WHEN** the skill template for any agent is rendered
- **THEN** it contains no instruction to estimate, count or pass token figures

#### Scenario: Agent declares the issue
- **WHEN** an agent follows the installed skill instructions and learns the issue it is working on
- **THEN** it calls `tokenpile bind` with that issue number
- **THEN** the call includes a `--note` and one or more `--tag` values from the vocabulary

#### Scenario: Not binding loses attribution, not measurement
- **WHEN** an agent never calls `tokenpile bind` during a session
- **THEN** the session's usage is still captured with complete token counts
- **THEN** those entries are unattributed and available for later assignment


### Requirement: Skill install location

Every supported agent (`claude-code`, `opencode`) SHALL install the tokenpile skill as a dedicated `SKILL.md` file, with a `name`/`description` YAML frontmatter, at the location that agent natively discovers skills (the Agent Skills spec layout: `<agent-skills-dir>/tokenpile/SKILL.md`). `Install` SHALL overwrite this file on repeat installs.

#### Scenario: Dedicated SKILL.md written for every agent
- **WHEN** `Install(agentName)` is called for any supported agent
- **THEN** a `SKILL.md` file with `name` and `description` frontmatter is written at that agent's native skill directory

#### Scenario: Supported set is exactly two agents
- **WHEN** the supported agent list is enumerated
- **THEN** it contains exactly `claude-code` and `opencode`

### Requirement: Unsupported agents are rejected

`Install`, `Uninstall` and `IsInstalled` SHALL return `ErrUnsupportedAgent` for any agent name outside the supported set, with no special-casing for names that were supported by a previous version. `codex` SHALL be rejected on the same terms as any unknown name.

#### Scenario: Installing a dropped agent fails
- **WHEN** `Install("codex")` is called
- **THEN** it returns `ErrUnsupportedAgent`
- **THEN** no file is written

#### Scenario: Dropped agent is not listed
- **WHEN** `tokenpile skill list` runs
- **THEN** the output names `claude-code` and `opencode` and does not name `codex`

### Requirement: Legacy skill install cleanup

`Install` and `Uninstall` SHALL clean up, on a best-effort basis, any install left by a previous tokenpile version that used a different location or format for that agent: a stale flat file SHALL be removed outright; a marked tokenpile block (`<!-- tokenpile:start -->` through `<!-- tokenpile:end -->`) inside a shared file (e.g. `AGENTS.md`) SHALL be stripped, preserving all other content in that file, and the file SHALL be removed entirely if nothing else remains. A failure to clean up a legacy location SHALL NOT block the current install/uninstall from succeeding.

#### Scenario: Legacy flat file removed on install
- **WHEN** a pre-migration flat skill file exists for an agent and `Install(agentName)` is called
- **THEN** the legacy flat file no longer exists
- **THEN** the new dedicated `SKILL.md` is written

#### Scenario: Legacy AGENTS.md block stripped, foreign content kept
- **WHEN** an agent's `AGENTS.md` contains user content plus a legacy tokenpile marked block
- **WHEN** `Install(agentName)` or `Uninstall(agentName)` is called
- **THEN** the tokenpile block is gone from `AGENTS.md` and the user content is intact

#### Scenario: Uninstall when not installed
- **WHEN** `Uninstall` is called for an agent with no installed skill
- **THEN** the call succeeds and reports that nothing was removed
