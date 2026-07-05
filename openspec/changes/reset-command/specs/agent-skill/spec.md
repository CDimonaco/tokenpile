## ADDED Requirements

### Requirement: Skill uninstall

The system SHALL provide `skill.Uninstall(agentName)` reversing `Install`. For agents with a dedicated skill file the file SHALL be removed. For agents sharing a file (e.g. `AGENTS.md`) only the marked tokenpile block (`<!-- tokenpile:start -->` through `<!-- tokenpile:end -->`) SHALL be removed, preserving all other content; if the file contains nothing else afterwards, the file SHALL be removed. Uninstalling a skill that is not installed SHALL succeed and report that nothing was removed.

#### Scenario: Dedicated skill file removed
- **WHEN** the claude-code skill is installed and `Uninstall("claude-code")` is called
- **THEN** the skill file no longer exists

#### Scenario: Shared file keeps foreign content
- **WHEN** an `AGENTS.md` contains user content plus the tokenpile marked block
- **WHEN** `Uninstall("codex")` is called
- **THEN** the tokenpile block is gone and the user content is intact

#### Scenario: Uninstall when not installed
- **WHEN** `Uninstall` is called for an agent with no installed skill
- **THEN** the call succeeds and reports that nothing was removed
