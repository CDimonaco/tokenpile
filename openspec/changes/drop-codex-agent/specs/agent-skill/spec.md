## ADDED Requirements

### Requirement: Unsupported agents are rejected

`Install`, `Uninstall` and `IsInstalled` SHALL return `ErrUnsupportedAgent` for any agent name outside the supported set, with no special-casing for names that were supported by a previous version. `codex` SHALL be rejected on the same terms as any unknown name.

#### Scenario: Installing a dropped agent fails
- **WHEN** `Install("codex")` is called
- **THEN** it returns `ErrUnsupportedAgent`
- **THEN** no file is written

#### Scenario: Dropped agent is not listed
- **WHEN** `tokenpile skill list` runs
- **THEN** the output names `claude-code` and `opencode` and does not name `codex`

## MODIFIED Requirements

### Requirement: Skill install location

Every supported agent (`claude-code`, `opencode`) SHALL install the tokenpile skill as a dedicated `SKILL.md` file, with a `name`/`description` YAML frontmatter, at the location that agent natively discovers skills (the Agent Skills spec layout: `<agent-skills-dir>/tokenpile/SKILL.md`). `Install` SHALL overwrite this file on repeat installs.

#### Scenario: Dedicated SKILL.md written for every agent
- **WHEN** `Install(agentName)` is called for any supported agent
- **THEN** a `SKILL.md` file with `name` and `description` frontmatter is written at that agent's native skill directory

#### Scenario: Supported set is exactly two agents
- **WHEN** the supported agent list is enumerated
- **THEN** it contains exactly `claude-code` and `opencode`
