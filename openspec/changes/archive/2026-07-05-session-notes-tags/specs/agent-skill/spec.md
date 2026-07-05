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
