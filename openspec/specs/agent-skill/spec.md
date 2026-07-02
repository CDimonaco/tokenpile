
### Requirement: Skill install subcommand

The system SHALL provide `tokenpile skill install --agent <name>` which writes an agent-specific skill file to the correct location for that agent. The skill file instructs the agent how to call `tokenpile log` after each LLM response.

Supported agents in v1:
- `claude-code`: writes to `~/.claude/skills/tokenpile.md`

#### Scenario: Install Claude Code skill
- **WHEN** `tokenpile skill install --agent claude-code`
- **THEN** the skill file is written to `~/.claude/skills/tokenpile.md`
- **THEN** the CLI prints "Skill installed for claude-code at ~/.claude/skills/tokenpile.md"

#### Scenario: Overwrite existing skill
- **WHEN** `tokenpile skill install --agent claude-code`
- **WHEN** a skill file already exists at the target path
- **THEN** the existing file is overwritten with the latest template
- **THEN** the CLI prints a warning: "Overwrote existing skill file"

#### Scenario: Unsupported agent
- **WHEN** `tokenpile skill install --agent unknown-agent`
- **THEN** the CLI exits with an error listing supported agent names

### Requirement: Skill templates embedded in binary

Agent skill file templates SHALL be embedded in the binary using Go's `embed` package from `internal/skill/templates/`. This ensures the skill content is always consistent with the CLI version installed.

#### Scenario: Skill template is embedded
- **WHEN** the binary is built
- **THEN** skill templates are accessible at runtime without external files

### Requirement: Skill file content

The Claude Code skill file SHALL instruct the agent to call `tokenpile log` with `--agent claude-code`, `--model <current-model>`, `--tokens-in <input-tokens>`, and `--tokens-out <output-tokens>` after each response. It SHALL document that `--repo` and `--issue` are required context that the user or agent must supply.

#### Scenario: Skill file provides correct CLI invocation
- **WHEN** the skill file is read by Claude Code
- **THEN** it instructs the agent to run `tokenpile log` with the required flags populated from the current session context

### Requirement: Skill list subcommand

The system SHALL provide `tokenpile skill list` which prints all supported agents and whether their skill file is currently installed.

#### Scenario: List shows installed and missing skills
- **WHEN** `tokenpile skill list`
- **WHEN** Claude Code skill is installed but no others
- **THEN** the output shows `claude-code: installed` and any other supported agents as `not installed`
