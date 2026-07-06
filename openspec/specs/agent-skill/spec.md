## MODIFIED Requirements

### Requirement: Skill file content

The skill template SHALL instruct the agent to call `tokenpile log` with `--note` and `--tag` flags in addition to the required flags. The note SHALL be a single line (max ~100 chars) summarizing what was done. Tags SHALL be chosen from a documented vocabulary: `refactor`, `debug`, `feature`, `test`, `docs`, `spike`, `review`.

#### Scenario: Skill file provides correct CLI invocation
- **WHEN** the skill template for any agent is rendered
- **THEN** the example `tokenpile log` invocation includes `--note` and `--tag` flags
- **THEN** the template documents the recommended tag vocabulary

#### Scenario: Agent passes note and tags on log
- **WHEN** an agent follows the installed skill instructions after completing work
- **THEN** the `tokenpile log` call includes a `--note` with a brief description
- **THEN** the call includes one or more `--tag` values from the vocabulary

### Requirement: Skill install location

Every supported agent (`claude-code`, `codex`, `opencode`) SHALL install the tokenpile skill as a dedicated `SKILL.md` file, with a `name`/`description` YAML frontmatter, at the location that agent natively discovers skills (the Agent Skills spec layout: `<agent-skills-dir>/tokenpile/SKILL.md`). `Install` SHALL overwrite this file on repeat installs.

#### Scenario: Dedicated SKILL.md written for every agent
- **WHEN** `Install(agentName)` is called for any supported agent
- **THEN** a `SKILL.md` file with `name` and `description` frontmatter is written at that agent's native skill directory

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
